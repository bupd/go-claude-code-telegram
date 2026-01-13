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
	Use:   "send",
	Short: "Send a message and wait for reply",
	Long:  "Read message from stdin, send to Telegram, wait for reply, print to stdout.",
	RunE:  runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)
}

func runSend(cmd *cobra.Command, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}
	message := strings.Join(lines, "\n")

	if message == "" {
		return fmt.Errorf("no message provided")
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
