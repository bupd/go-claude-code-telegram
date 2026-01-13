package session

import (
	"sync"
	"time"

	"github.com/bupd/go-claude-code-telegram/internal/config"
)

type PendingMessage struct {
	ID         string
	TgMsgID    int
	Content    string
	ResponseCh chan string
	CreatedAt  time.Time
}

type ChatIDCapture struct {
	ResponseCh chan int64
}

type Manager struct {
	config        *config.Config
	pending       map[int64][]*PendingMessage // keyed by chat_id
	queuedMsgs    map[int64][]string          // messages sent when no pending
	chatIDCapture *ChatIDCapture              // pending chat ID capture request
	mu            sync.RWMutex
	idSeq         int64
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:     cfg,
		pending:    make(map[int64][]*PendingMessage),
		queuedMsgs: make(map[int64][]string),
	}
}

func (m *Manager) Config() *config.Config {
	return m.config
}

func (m *Manager) AddPending(chatID int64, tgMsgID int, content string) *PendingMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.idSeq++
	pm := &PendingMessage{
		ID:         string(rune(m.idSeq)),
		TgMsgID:    tgMsgID,
		Content:    content,
		ResponseCh: make(chan string, 1),
		CreatedAt:  time.Now(),
	}

	m.pending[chatID] = append(m.pending[chatID], pm)
	return pm
}

func (m *Manager) HandleReply(chatID int64, replyToMsgID int, reply string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.pending[chatID]
	if !exists || len(queue) == 0 {
		return false
	}

	var matched *PendingMessage
	var matchedIdx int

	if replyToMsgID > 0 {
		for i, pm := range queue {
			if pm.TgMsgID == replyToMsgID {
				matched = pm
				matchedIdx = i
				break
			}
		}
	}

	if matched == nil {
		matched = queue[0]
		matchedIdx = 0
	}

	matched.ResponseCh <- reply
	close(matched.ResponseCh)

	m.pending[chatID] = append(queue[:matchedIdx], queue[matchedIdx+1:]...)
	if len(m.pending[chatID]) == 0 {
		delete(m.pending, chatID)
	}

	return true
}

func (m *Manager) RemovePending(chatID int64, pm *PendingMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.pending[chatID]
	if !exists {
		return
	}

	for i, p := range queue {
		if p.ID == pm.ID {
			m.pending[chatID] = append(queue[:i], queue[i+1:]...)
			if len(m.pending[chatID]) == 0 {
				delete(m.pending, chatID)
			}
			return
		}
	}
}

func (m *Manager) HasPendingForChat(chatID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	queue, exists := m.pending[chatID]
	return exists && len(queue) > 0
}

func (m *Manager) FindSessionByName(name string) *config.SessionConfig {
	return m.config.FindSessionByName(name)
}

func (m *Manager) FindSessionByWorkDir(workDir string) *config.SessionConfig {
	return m.config.FindSessionByWorkDir(workDir)
}

func (m *Manager) QueueMessage(chatID int64, text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queuedMsgs[chatID] = append(m.queuedMsgs[chatID], text)
}

func (m *Manager) PopQueuedMessages(chatID int64) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	msgs := m.queuedMsgs[chatID]
	delete(m.queuedMsgs, chatID)
	return msgs
}

func (m *Manager) StartChatIDCapture() *ChatIDCapture {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chatIDCapture = &ChatIDCapture{
		ResponseCh: make(chan int64, 1),
	}
	return m.chatIDCapture
}

func (m *Manager) CancelChatIDCapture() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chatIDCapture = nil
}

func (m *Manager) TryCaptureChat(chatID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.chatIDCapture != nil {
		m.chatIDCapture.ResponseCh <- chatID
		close(m.chatIDCapture.ResponseCh)
		m.chatIDCapture = nil
		return true
	}
	return false
}
