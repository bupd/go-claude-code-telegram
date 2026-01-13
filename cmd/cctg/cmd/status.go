package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
	"github.com/bupd/go-claude-code-telegram/internal/ipc"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if daemon is running",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	client := ipc.NewClient(config.GetSocketPath())

	if client.IsRunning() {
		fmt.Println("daemon is running")
	} else {
		fmt.Println("daemon is not running")
	}
	return nil
}
