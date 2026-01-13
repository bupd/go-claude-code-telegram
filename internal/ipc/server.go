package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

type RequestHandler func(req *Request) *Response

type Server struct {
	socketPath string
	listener   net.Listener
	handler    RequestHandler
}

func NewServer(socketPath string, handler RequestHandler) *Server {
	return &Server{
		socketPath: socketPath,
		handler:    handler,
	}
}

func (s *Server) Start(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0700); err != nil {
		return fmt.Errorf("creating socket directory: %w", err)
	}

	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing existing socket: %w", err)
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("creating unix socket: %w", err)
	}
	s.listener = listener

	if err := os.Chmod(s.socketPath, 0600); err != nil {
		return fmt.Errorf("setting socket permissions: %w", err)
	}

	go s.acceptLoop(ctx)

	return nil
}

func (s *Server) acceptLoop(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		s.sendResponse(conn, &Response{
			Success: false,
			Error:   "invalid request format",
		})
		return
	}

	resp := s.handler(&req)
	s.sendResponse(conn, resp)
}

func (s *Server) sendResponse(conn net.Conn, resp *Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)
}

func (s *Server) Stop() error {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.socketPath)
	return nil
}

func (s *Server) SocketPath() string {
	return s.socketPath
}

func DefaultSocketPath() string {
	return config.GetSocketPath()
}
