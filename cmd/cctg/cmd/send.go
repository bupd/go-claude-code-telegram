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

var sendCmd = &cobra.Command{
	Use:   "send [message]",
	Short: "Send a message and wait for reply",
	Long: `Send a message to Telegram and wait for user reply.

Message can be provided as:
  - Arguments: cctg send "your message here"
  - Stdin: echo "your message" | cctg send

Session is auto-detected from working directory, or specify with --session.`,
	Example: `  cctg send "Should I proceed with the refactor?"
  cctg send --session myproject "Deploy to production?"
  echo "Review this change?" | cctg send`,
	RunE: runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)
}

func runSend(cmd *cobra.Command, args []string) error {
	var message string

	if len(args) > 0 {
		message = strings.Join(args, " ")
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			message = strings.Join(lines, "\n")
		}
	}

	if message == "" {
		return fmt.Errorf("message required: cctg send \"your message\" or echo \"message\" | cctg send")
	}

	client := ipc.NewClient(config.GetSocketPath())

	if !client.IsRunning() {
		fmt.Println("user didn't reply go ahead with caution, don't make huge refactor, check what you are doing")
		return nil
	}

	workDir, _ := os.Getwd()

	req := &ipc.Request{
		Type:    ipc.RequestTypeSend,
		Session: sessionArg,
		Message: message,
		Timeout: timeoutArg,
		WorkDir: workDir,
	}

	resp, err := client.Send(req)
	if err != nil {
		fmt.Println("user didn't reply go ahead with caution, don't make huge refactor, check what you are doing")
		return nil
	}

	if !resp.Success {
		return fmt.Errorf("send failed: %s", resp.Error)
	}

	fmt.Println(resp.Reply)
	return nil
}
