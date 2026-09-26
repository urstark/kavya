package bot

import (
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

type BurstCallback func(c tele.Context, aggregatedText string)

type BurstItem struct {
	Context   tele.Context
	Messages  []string
	Timer     *time.Timer
}

type BurstManager struct {
	mu         sync.Mutex
	delay      time.Duration
	pending    map[int64]*BurstItem
	onComplete BurstCallback
}

func NewBurstManager(delaySec float64, onComplete BurstCallback) *BurstManager {
	return &BurstManager{
		delay:      time.Duration(delaySec * float64(time.Second)),
		pending:    make(map[int64]*BurstItem),
		onComplete: onComplete,
	}
}

func (b *BurstManager) AddMessage(c tele.Context, text string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	chatID := c.Chat().ID
	item, exists := b.pending[chatID]

	if !exists {
		item = &BurstItem{
			Context:  c,
			Messages: []string{text},
		}
		item.Timer = time.AfterFunc(b.delay, func() {
			b.flush(chatID)
		})
		b.pending[chatID] = item
	} else {
		item.Timer.Stop()
		item.Messages = append(item.Messages, text)
		item.Context = c // Update context to the latest message
		item.Timer = time.AfterFunc(b.delay, func() {
			b.flush(chatID)
		})
	}
}

func (b *BurstManager) flush(chatID int64) {
	b.mu.Lock()
	item, exists := b.pending[chatID]
	if !exists {
		b.mu.Unlock()
		return
	}
	delete(b.pending, chatID)
	b.mu.Unlock()

	combined := strings.Join(item.Messages, "\n")
	b.onComplete(item.Context, combined)
}
