package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/bupd/go-claude-code-telegram/internal/config"
	"github.com/bupd/go-claude-code-telegram/internal/session"
)

const MaxMessageLength = 4096

type Bot struct {
	api      *tgbotapi.BotAPI
	config   *config.Config
	sessions *session.Manager
}

func NewBot(cfg *config.Config, sessions *session.Manager) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("creating bot api: %w", err)
	}

	return &Bot{
		api:      api,
		config:   cfg,
		sessions: sessions,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.notifyAllSessions("cctg daemon started")

	for {
		select {
		case <-ctx.Done():
			b.notifyAllSessions("cctg daemon stopped")
			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			b.handleMessage(update.Message)
		}
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	if !b.isAllowedUser(msg.From.ID) {
		return
	}

	chatID := msg.Chat.ID

	if b.sessions.HasPendingForChat(chatID) {
		replyToMsgID := 0
		if msg.ReplyToMessage != nil {
			replyToMsgID = msg.ReplyToMessage.MessageID
		}
		b.sessions.HandleReply(chatID, replyToMsgID, msg.Text)
	} else {
		b.sessions.QueueMessage(chatID, msg.Text)
	}
}

func (b *Bot) isAllowedUser(userID int64) bool {
	for _, allowed := range b.config.Telegram.AllowedUsers {
		if allowed == userID {
			return true
		}
	}
	return false
}

func (b *Bot) SendMessage(chatID int64, text string) (int, error) {
	if len(text) > MaxMessageLength {
		return 0, fmt.Errorf("message exceeds %d character limit", MaxMessageLength)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	sent, err := b.api.Send(msg)
	if err != nil {
		return 0, fmt.Errorf("sending message: %w", err)
	}
	return sent.MessageID, nil
}

func (b *Bot) notifyAllSessions(text string) {
	for _, sess := range b.config.Sessions {
		msg := tgbotapi.NewMessage(sess.ChatID, text)
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("failed to notify chat %d: %v", sess.ChatID, err)
		}
	}
}

func (b *Bot) NotifyTimeout(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "timeout: no reply received")
	b.api.Send(msg)
}
