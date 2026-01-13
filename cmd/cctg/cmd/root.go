package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	sessionArg string
	timeoutArg int
)

var rootCmd = &cobra.Command{
	Use:   "cctg",
	Short: "Claude Code Telegram Bot",
	Long:  "A Telegram bot that bridges Claude Code CLI with Telegram users.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&sessionArg, "session", "", "session name")
	rootCmd.PersistentFlags().IntVar(&timeoutArg, "timeout", 0, "timeout in seconds")
}
