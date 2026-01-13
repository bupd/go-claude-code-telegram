package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
	"github.com/bupd/go-claude-code-telegram/internal/ipc"
	"github.com/bupd/go-claude-code-telegram/internal/session"
	"github.com/bupd/go-claude-code-telegram/internal/telegram"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Telegram bot daemon",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	sessions := session.NewManager(cfg)

	bot, err := telegram.NewBot(cfg, sessions)
	if err != nil {
		return fmt.Errorf("creating bot: %w", err)
	}

	server := ipc.NewServer(config.GetSocketPath(), func(req *ipc.Request) *ipc.Response {
		return handleIPCRequest(req, cfg, sessions, bot)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("starting ipc server: %w", err)
	}
	defer server.Stop()

	log.Printf("daemon started, socket: %s", server.SocketPath())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- bot.Start(ctx)
	}()

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %s", sig)
		cancel()
		return nil
	case err := <-errCh:
		return err
	}
}

func handleIPCRequest(req *ipc.Request, cfg *config.Config, sessions *session.Manager, bot *telegram.Bot) *ipc.Response {
	switch req.Type {
	case ipc.RequestTypeGetChatID:
		return handleGetChatID(req, sessions)
	case ipc.RequestTypeSend:
		return handleSend(req, cfg, sessions, bot)
	default:
		return &ipc.Response{Success: false, Error: "unknown request type"}
	}
}

func handleGetChatID(req *ipc.Request, sessions *session.Manager) *ipc.Response {
	capture := sessions.StartChatIDCapture()

	timeout := 60
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	select {
	case chatID := <-capture.ResponseCh:
		return &ipc.Response{Success: true, ChatID: chatID}
	case <-time.After(time.Duration(timeout) * time.Second):
		sessions.CancelChatIDCapture()
		return &ipc.Response{Success: false, Error: "timeout waiting for message"}
	}
}

func handleSend(req *ipc.Request, cfg *config.Config, sessions *session.Manager, bot *telegram.Bot) *ipc.Response {

	var sess *config.SessionConfig
	if req.Session != "" {
		sess = cfg.FindSessionByName(req.Session)
	} else if req.WorkDir != "" {
		sess = cfg.FindSessionByWorkDir(req.WorkDir)
	}

	if sess == nil {
		return &ipc.Response{Success: false, Error: "session not found"}
	}

	queued := sessions.PopQueuedMessages(sess.ChatID)

	msgID, err := bot.SendMessage(sess.ChatID, req.Message)
	if err != nil {
		return &ipc.Response{Success: false, Error: err.Error()}
	}

	pending := sessions.AddPending(sess.ChatID, msgID, req.Message)

	timeout := cfg.Timeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	select {
	case reply := <-pending.ResponseCh:
		finalReply := combineMessages(queued, reply)
		return &ipc.Response{Success: true, Reply: finalReply}
	case <-time.After(time.Duration(timeout) * time.Second):
		sessions.RemovePending(sess.ChatID, pending)
		bot.NotifyTimeout(sess.ChatID)
		if len(queued) > 0 {
			return &ipc.Response{Success: true, Reply: strings.Join(queued, "\n")}
		}
		return &ipc.Response{
			Success: true,
			Reply:   "user didn't reply go ahead with caution, don't make huge refactor, check what you are doing",
		}
	}
}

func combineMessages(queued []string, reply string) string {
	if len(queued) == 0 {
		return reply
	}
	parts := append(queued, reply)
	return strings.Join(parts, "\n")
}
