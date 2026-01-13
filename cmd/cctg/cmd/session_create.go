package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
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

Required information:
  --name         Unique session identifier (required)
  --chat-id      Telegram chat ID where messages are sent (required)
  --working-dir  Directory path associated with this session (required)

If flags are not provided, prompts interactively for values.

Examples:
  # Create with flags (non-interactive)
  cctg session create --name api --chat-id 123456789 --working-dir /path/to/project

  # Create interactively
  cctg session create`,
	RunE: runSessionCreate,
}

func init() {
	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCreateCmd.Flags().StringVar(&createName, "name", "", "session name (required)")
	sessionCreateCmd.Flags().Int64Var(&createChatID, "chat-id", 0, "telegram chat ID (required)")
	sessionCreateCmd.Flags().StringVar(&createWorkingDir, "working-dir", "", "working directory path (required)")
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
		fmt.Print("Chat ID: ")
		chatIDStr, _ := reader.ReadString('\n')
		chatIDStr = strings.TrimSpace(chatIDStr)
		chatID, err = strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chat ID: %w", err)
		}
	}
	if chatID == 0 {
		return fmt.Errorf("chat ID is required")
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
