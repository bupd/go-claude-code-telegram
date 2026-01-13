package cmd

import (
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long: `Manage cctg sessions. Sessions map working directories to Telegram chats.

Each session has:
  - name: unique identifier for the session
  - chat_id: Telegram chat where messages are sent
  - working_dir: directory that triggers this session`,
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}
