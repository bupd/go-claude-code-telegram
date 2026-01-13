package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:        "list",
	Short:      "List configured sessions (alias for 'session list')",
	Deprecated: "use 'cctg session list' instead",
	RunE:       runSessionList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
