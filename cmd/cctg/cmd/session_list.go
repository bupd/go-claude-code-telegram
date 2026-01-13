package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured sessions",
	Long: `List all sessions configured in cctg.

Each session displays:
  - name: session identifier
  - chat_id: Telegram chat ID
  - working_dir: associated working directory

Example:
  cctg session list`,
	RunE: runSessionList,
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
}

func runSessionList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Sessions) == 0 {
		fmt.Println("no sessions configured")
		return nil
	}

	for _, sess := range cfg.Sessions {
		fmt.Printf("%s\n  chat_id: %d\n  working_dir: %s\n", sess.Name, sess.ChatID, sess.WorkingDir)
	}
	return nil
}
