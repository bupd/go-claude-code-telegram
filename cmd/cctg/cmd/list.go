package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured sessions",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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
