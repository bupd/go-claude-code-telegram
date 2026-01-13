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
	editName       string
	editChatID     int64
	editWorkingDir string
)

var sessionEditCmd = &cobra.Command{
	Use:   "edit <session-name>",
	Short: "Edit an existing session",
	Long: `Edit an existing session configuration.

Arguments:
  session-name  Name of the session to edit (required)

Flags (optional):
  --name         New session name
  --chat-id      New Telegram chat ID
  --working-dir  New working directory path

If no flags provided, prompts interactively showing current values.
Press Enter to keep current value.

Examples:
  # Edit with flags
  cctg session edit api --chat-id 987654321

  # Edit interactively
  cctg session edit api

  # Rename session
  cctg session edit api --name api-v2`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionEdit,
}

func init() {
	sessionCmd.AddCommand(sessionEditCmd)
	sessionEditCmd.Flags().StringVar(&editName, "name", "", "new session name")
	sessionEditCmd.Flags().Int64Var(&editChatID, "chat-id", 0, "new telegram chat ID")
	sessionEditCmd.Flags().StringVar(&editWorkingDir, "working-dir", "", "new working directory path")
}

func runSessionEdit(cmd *cobra.Command, args []string) error {
	sessionName := args[0]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	session := cfg.FindSessionByName(sessionName)
	if session == nil {
		return fmt.Errorf("session %q not found", sessionName)
	}

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("chat-id") || cmd.Flags().Changed("working-dir")

	if flagsProvided {
		if cmd.Flags().Changed("name") {
			if editName != sessionName && cfg.FindSessionByName(editName) != nil {
				return fmt.Errorf("session %q already exists", editName)
			}
			session.Name = editName
		}
		if cmd.Flags().Changed("chat-id") {
			session.ChatID = editChatID
		}
		if cmd.Flags().Changed("working-dir") {
			session.WorkingDir = editWorkingDir
		}
	} else {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Name [%s]: ", session.Name)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			if input != session.Name && cfg.FindSessionByName(input) != nil {
				return fmt.Errorf("session %q already exists", input)
			}
			session.Name = input
		}

		fmt.Printf("Chat ID [%d]: ", session.ChatID)
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			chatID, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chat ID: %w", err)
			}
			session.ChatID = chatID
		}

		fmt.Printf("Working directory [%s]: ", session.WorkingDir)
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			session.WorkingDir = input
		}
	}

	if err := cfg.Save(""); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("session %q updated\n", session.Name)
	return nil
}
