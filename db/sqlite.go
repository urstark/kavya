package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"

	"kavya/ai"
)

type SQLiteStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_chat_messages ON messages(chat_id, id);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize sqlite schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) SaveMessage(chatID int64, role, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO messages (chat_id, role, content) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, chatID, role, content)
	return err
}

func (s *SQLiteStore) GetHistory(chatID int64, limit int) ([]ai.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}

	query := `
	SELECT role, content FROM (
		SELECT id, role, content FROM messages 
		WHERE chat_id = ? 
		ORDER BY id DESC 
		LIMIT ?
	) ORDER BY id ASC;
	`

	rows, err := s.db.Query(query, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ai.Message
	for rows.Next() {
		var msg ai.Message
		if err := rows.Scan(&msg.Role, &msg.Content); err != nil {
			return nil, err
		}
		history = append(history, msg)
	}

	return history, rows.Err()
}

func (s *SQLiteStore) ClearHistory(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM messages WHERE chat_id = ?`
	_, err := s.db.Exec(query, chatID)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
