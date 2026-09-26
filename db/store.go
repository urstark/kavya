package db

import (
	"kavya/ai"
)

type Store interface {
	SaveMessage(chatID int64, role, content string) error
	GetHistory(chatID int64, limit int) ([]ai.Message, error)
	ClearHistory(chatID int64) error
	Close() error
}
