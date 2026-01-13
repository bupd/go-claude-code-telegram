package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

var (
	initToken       string
	initUserID      int64
	initUsername    string
	initSessionName string
	initChatID      int64
	initWorkingDir  string
	initTimeout     int
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
	initCmd.Flags().Int64Var(&initUserID, "user-id", 0, "Your Telegram user ID (numeric)")
	initCmd.Flags().StringVar(&initUsername, "username", "", "Your Telegram username (e.g., @username)")
	initCmd.Flags().StringVar(&initSessionName, "session-name", "", "Session name")
	initCmd.Flags().Int64Var(&initChatID, "chat-id", 0, "Telegram chat ID")
	initCmd.Flags().StringVar(&initWorkingDir, "working-dir", "", "Working directory for session")
	initCmd.Flags().IntVar(&initTimeout, "timeout", 0, "Timeout in seconds (default 300)")
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
	if userID == 0 && initUsername != "" {
		var err error
		userID, err = resolveUsername(token, initUsername)
		if err != nil {
			return err
		}
	}

	if userID == 0 {
		input := prompt(reader, "Your Telegram user ID or @username")
		if strings.HasPrefix(input, "@") {
			fmt.Println("\nTo resolve username, send /start to your bot first.")
			prompt(reader, "Press Enter after sending /start to the bot")
			var err error
			userID, err = resolveUsername(token, input)
			if err != nil {
				return err
			}
		} else {
			var err error
			userID, err = strconv.ParseInt(input, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid user ID: %s", input)
			}
		}
	}

	if userID == 0 {
		return fmt.Errorf("user ID is required")
	}
	fmt.Printf("User ID: %d\n", userID)

	sessionName := initSessionName
	if sessionName == "" {
		sessionName = prompt(reader, "Session name (e.g., myproject)")
	}
	if sessionName == "" {
		sessionName = "default"
	}

	chatID := initChatID
	if chatID == 0 {
		fmt.Println("\nChat ID options:")
		fmt.Printf("  - Press Enter to use private chat (%d)\n", userID)
		fmt.Println("  - Type 'group' to detect from a group (add bot to group and send a message first)")
		fmt.Println("  - Enter a chat ID directly")
		input := prompt(reader, "Chat ID")
		if input == "" {
			chatID = userID
		} else if strings.ToLower(input) == "group" {
			var err error
			chatID, err = detectGroupChat(token)
			if err != nil {
				return err
			}
			fmt.Printf("Detected group chat ID: %d\n", chatID)
		} else {
			var err error
			chatID, err = strconv.ParseInt(input, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chat ID: %s", input)
			}
		}
	}

	workingDir := initWorkingDir
	if workingDir == "" {
		cwd, _ := os.Getwd()
		workingDir = prompt(reader, fmt.Sprintf("Working directory [%s]", cwd))
		if workingDir == "" {
			workingDir = cwd
		}
	}

	timeout := initTimeout
	if timeout == 0 {
		input := prompt(reader, "Timeout in seconds [300]")
		if input == "" {
			timeout = 300
		} else {
			var err error
			timeout, err = strconv.Atoi(input)
			if err != nil {
				return fmt.Errorf("invalid timeout: %s", input)
			}
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

func resolveUsername(token, username string) (int64, error) {
	username = strings.TrimPrefix(username, "@")

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", token)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("calling telegram API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				From struct {
					ID       int64  `json:"id"`
					Username string `json:"username"`
				} `json:"from"`
			} `json:"message"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("parsing response: %w", err)
	}

	if !result.OK {
		return 0, fmt.Errorf("telegram API error")
	}

	for _, update := range result.Result {
		if strings.EqualFold(update.Message.From.Username, username) {
			return update.Message.From.ID, nil
		}
	}

	return 0, fmt.Errorf("username @%s not found in recent messages - send /start to the bot first", username)
}

func detectGroupChat(token string) (int64, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", token)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("calling telegram API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				Chat struct {
					ID   int64  `json:"id"`
					Type string `json:"type"`
				} `json:"chat"`
			} `json:"message"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("parsing response: %w", err)
	}

	if !result.OK {
		return 0, fmt.Errorf("telegram API error")
	}

	for _, update := range result.Result {
		chatType := update.Message.Chat.Type
		if chatType == "group" || chatType == "supergroup" {
			return update.Message.Chat.ID, nil
		}
	}

	return 0, fmt.Errorf("no group chat found - add bot to a group and send a message first")
}

func prompt(reader *bufio.Reader, question string) string {
	fmt.Printf("%s: ", question)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
