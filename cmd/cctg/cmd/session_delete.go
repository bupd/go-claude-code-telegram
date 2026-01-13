package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

var deleteForce bool

var sessionDeleteCmd = &cobra.Command{
	Use:   "delete <session-name>",
	Short: "Delete a session",
	Long: `Delete an existing session from the configuration.

Arguments:
  session-name  Name of the session to delete (required)

Flags:
  --force, -f   Skip confirmation prompt

Prompts for confirmation before deleting unless --force is used.

Examples:
  # Delete with confirmation
  cctg session delete api

  # Delete without confirmation
  cctg session delete api --force`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionDelete,
}

func init() {
	sessionCmd.AddCommand(sessionDeleteCmd)
	sessionDeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompt")
}

func runSessionDelete(cmd *cobra.Command, args []string) error {
	sessionName := args[0]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	idx := -1
	for i, s := range cfg.Sessions {
		if s.Name == sessionName {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("session %q not found", sessionName)
	}

	if !deleteForce {
		fmt.Printf("Delete session %q? [y/N]: ", sessionName)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("cancelled")
			return nil
		}
	}

	cfg.Sessions = append(cfg.Sessions[:idx], cfg.Sessions[idx+1:]...)

	if err := cfg.Save(""); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("session %q deleted\n", sessionName)
	return nil
}
