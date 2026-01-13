package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

var (
	initToken      string
	initUserID     int64
	initSessionName string
	initChatID     int64
	initWorkingDir string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Create config directory and files interactively or via flags.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initToken, "token", "", "Telegram bot token")
	initCmd.Flags().Int64Var(&initUserID, "user-id", 0, "Your Telegram user ID")
	initCmd.Flags().StringVar(&initSessionName, "session-name", "", "Session name")
	initCmd.Flags().Int64Var(&initChatID, "chat-id", 0, "Telegram chat ID")
	initCmd.Flags().StringVar(&initWorkingDir, "working-dir", "", "Working directory for session")
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	token := initToken
	if token == "" {
		token = prompt(reader, "Bot token (from @BotFather)")
	}
	if token == "" {
		return fmt.Errorf("bot token is required")
	}

	userID := initUserID
	if userID == 0 {
		userID = promptInt64(reader, "Your Telegram user ID (from @userinfobot)")
	}
	if userID == 0 {
		return fmt.Errorf("user ID is required")
	}

	sessionName := initSessionName
	if sessionName == "" {
		sessionName = prompt(reader, "Session name (e.g., myproject)")
	}
	if sessionName == "" {
		sessionName = "default"
	}

	chatID := initChatID
	if chatID == 0 {
		chatID = promptInt64(reader, "Chat ID (where bot sends messages)")
	}
	if chatID == 0 {
		chatID = userID
	}

	workingDir := initWorkingDir
	if workingDir == "" {
		cwd, _ := os.Getwd()
		workingDir = prompt(reader, fmt.Sprintf("Working directory [%s]", cwd))
		if workingDir == "" {
			workingDir = cwd
		}
	}

	configDir := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	envPath := filepath.Join(configDir, ".env")
	envContent := fmt.Sprintf("TELEGRAM_BOT_TOKEN=%s\n", token)
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		return fmt.Errorf("writing .env: %w", err)
	}
	fmt.Printf("Created %s\n", envPath)

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := fmt.Sprintf(`telegram:
  allowed_users:
    - %d

timeout: 300

sessions:
  - name: "%s"
    chat_id: %d
    working_dir: "%s"
`, userID, sessionName, chatID, workingDir)

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("writing config.yaml: %w", err)
	}
	fmt.Printf("Created %s\n", configPath)

	fmt.Println("\nConfiguration complete. Run 'cctg serve' to start the daemon.")
	return nil
}

func prompt(reader *bufio.Reader, question string) string {
	fmt.Printf("%s: ", question)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptInt64(reader *bufio.Reader, question string) int64 {
	input := prompt(reader, question)
	if input == "" {
		return 0
	}
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
