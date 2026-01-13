package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
	"github.com/bupd/go-claude-code-telegram/internal/ipc"
)

var (
	createName       string
	createChatID     int64
	createWorkingDir string
)

var sessionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new session",
	Long: `Create a new session that maps a working directory to a Telegram chat.

When cctg send is called from a directory, it automatically finds the session
with matching working-dir and sends messages to that chat. This allows
different projects to use different Telegram chats without specifying --session.

Flags:
  --name         Unique session identifier (required)
  --chat-id      Telegram chat ID (auto-detected if daemon running)
  --working-dir  Project directory that uses this session (required)

If --chat-id is not provided and daemon is running, send a message to the
bot in the target chat to auto-detect the chat ID.

Examples:
  # Create session for a project (daemon must be running for chat-id detection)
  cctg session create --name api --working-dir /home/user/api-project
  # Then from /home/user/api-project, "cctg send" goes to this chat

  # Create with explicit chat ID
  cctg session create --name api --chat-id 123456789 --working-dir /path/to/project`,
	RunE: runSessionCreate,
}

func init() {
	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCreateCmd.Flags().StringVar(&createName, "name", "", "session name (required)")
	sessionCreateCmd.Flags().Int64Var(&createChatID, "chat-id", 0, "telegram chat ID (auto-detected if not provided)")
	sessionCreateCmd.Flags().StringVar(&createWorkingDir, "working-dir", "", "project directory that uses this session (required)")
}

func runSessionCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	name := createName
	if name == "" {
		fmt.Print("Session name: ")
		name, _ = reader.ReadString('\n')
		name = strings.TrimSpace(name)
	}
	if name == "" {
		return fmt.Errorf("session name is required")
	}

	if cfg.FindSessionByName(name) != nil {
		return fmt.Errorf("session %q already exists", name)
	}

	chatID := createChatID
	if chatID == 0 {
		client := ipc.NewClient(config.GetSocketPath())
		if !client.IsRunning() {
			return fmt.Errorf("daemon not running. start with: cctg serve")
		}

		fmt.Println("send a message to the bot in the chat you want to link...")
		resp, err := client.Send(&ipc.Request{
			Type:    ipc.RequestTypeGetChatID,
			Timeout: 60,
		})
		if err != nil {
			return fmt.Errorf("getting chat ID: %w", err)
		}
		if !resp.Success {
			return fmt.Errorf("getting chat ID: %s", resp.Error)
		}
		chatID = resp.ChatID
		fmt.Printf("captured chat ID: %d\n", chatID)
	}

	workingDir := createWorkingDir
	if workingDir == "" {
		fmt.Print("Working directory: ")
		workingDir, _ = reader.ReadString('\n')
		workingDir = strings.TrimSpace(workingDir)
	}
	if workingDir == "" {
		return fmt.Errorf("working directory is required")
	}

	cfg.Sessions = append(cfg.Sessions, config.SessionConfig{
		Name:       name,
		ChatID:     chatID,
		WorkingDir: workingDir,
	})

	if err := cfg.Save(""); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("session %q created\n", name)
	return nil
}
