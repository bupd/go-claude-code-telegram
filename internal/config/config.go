package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

type Config struct {
	Telegram TelegramConfig  `mapstructure:"telegram"`
	Timeout  int             `mapstructure:"timeout"`
	Sessions []SessionConfig `mapstructure:"sessions"`
}

type TelegramConfig struct {
	BotToken     string  `mapstructure:"bot_token"`
	AllowedUsers []int64 `mapstructure:"allowed_users"`
}

type SessionConfig struct {
	Name       string `mapstructure:"name"`
	ChatID     int64  `mapstructure:"chat_id"`
	WorkingDir string `mapstructure:"working_dir"`
}

const (
	DefaultTimeout    = 300
	DefaultConfigDir  = ".config/cctg"
	DefaultConfigFile = "config.yaml"
	DefaultEnvFile    = ".env"
	DefaultSocketFile = "cctg.sock"
)

func Load(configPath string) (*Config, error) {
	if err := loadEnvFile(); err != nil {
		return nil, fmt.Errorf("loading env file: %w", err)
	}

	v := viper.New()
	v.SetConfigType("yaml")

	v.SetDefault("timeout", DefaultTimeout)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath(getConfigDir())
	}

	v.AutomaticEnv()
	v.BindEnv("telegram.bot_token", "TELEGRAM_BOT_TOKEN")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found")
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.Telegram.BotToken == "" {
		return nil, fmt.Errorf("telegram.bot_token is required")
	}

	return &cfg, nil
}

func loadEnvFile() error {
	envPath := filepath.Join(getConfigDir(), DefaultEnvFile)
	if _, err := os.Stat(envPath); err == nil {
		return godotenv.Load(envPath)
	}
	return nil
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, DefaultConfigDir)
}

func GetSocketPath() string {
	return filepath.Join(getConfigDir(), DefaultSocketFile)
}

func (c *Config) FindSessionByName(name string) *SessionConfig {
	for i := range c.Sessions {
		if c.Sessions[i].Name == name {
			return &c.Sessions[i]
		}
	}
	return nil
}

func (c *Config) FindSessionByWorkDir(workDir string) *SessionConfig {
	for i := range c.Sessions {
		if c.Sessions[i].WorkingDir == workDir {
			return &c.Sessions[i]
		}
	}
	return nil
}
