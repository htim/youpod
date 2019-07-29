package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"sync"
)

type LogAppender struct {
	bot *YouPod

	mu    sync.RWMutex
	chats []int64
}

func NewAppender(bot *YouPod) *LogAppender {
	return &LogAppender{
		bot:   bot,
		mu:    sync.RWMutex{},
		chats: make([]int64, 0),
	}
}

func (a *LogAppender) RegisterChat(id int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.chats = append(a.chats, id)
}

func (a *LogAppender) Write(p []byte) (n int, err error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, chat := range a.chats {
		msg := tgbotapi.NewMessage(chat, string(p))
		if _, err := a.bot.api.Send(msg); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}
