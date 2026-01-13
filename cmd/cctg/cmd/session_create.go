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
	Long: `Create a new session that maps a name to a Telegram chat.

Use "cctg send --session <name>" to send messages to a specific session.
Alternatively, working-dir enables auto-detection based on current directory.

Flags:
  --name         Unique session identifier (required)
  --chat-id      Telegram chat ID (auto-detected if daemon running)
  --working-dir  Project directory for auto-detection (optional)

If --chat-id is not provided and daemon is running, send a message to the
bot in the target chat to auto-detect the chat ID.

Examples:
  # Create session (daemon must be running for chat-id detection)
  cctg session create --name api --working-dir /home/user/api-project

  # Send using session name (recommended)
  cctg send --session api "message"

  # Or auto-detect from working directory
  cd /home/user/api-project && cctg send "message"`,
	RunE: runSessionCreate,
}

func init() {
	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCreateCmd.Flags().StringVar(&createName, "name", "", "session name (required)")
	sessionCreateCmd.Flags().Int64Var(&createChatID, "chat-id", 0, "telegram chat ID (auto-detected if not provided)")
	sessionCreateCmd.Flags().StringVar(&createWorkingDir, "working-dir", "", "project directory for auto-detection (optional)")
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
