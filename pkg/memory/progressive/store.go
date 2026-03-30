package progressive

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type MemoryLayer int

const (
	LayerSearch   MemoryLayer = 1
	LayerTimeline MemoryLayer = 2
	LayerDetail   MemoryLayer = 3
)

type MemoryEntry struct {
	ID        string                 `json:"id"`
	Layer     MemoryLayer            `json:"layer"`
	SessionID string                 `json:"session_id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Summary   string                 `json:"summary"`
	Content   string                 `json:"content,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsPrivate bool                   `json:"is_private"`
}

type SearchResult struct {
	Entries    []MemoryEntry `json:"entries"`
	TotalFound int           `json:"total_found"`
	TokensUsed int           `json:"tokens_used"`
	Layer      MemoryLayer   `json:"layer_retrieved"`
}

type ProgressiveMemoryStore struct {
	db *sql.DB
}

func NewProgressiveMemoryStore(workspacePath string) (*ProgressiveMemoryStore, error) {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".heron", "memory", "progressive.db")
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	store := &ProgressiveMemoryStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *ProgressiveMemoryStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS memory_entries (
			id TEXT PRIMARY KEY,
			layer INTEGER NOT NULL,
			session_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			type TEXT NOT NULL,
			summary TEXT NOT NULL,
			content TEXT,
			tags TEXT,
			metadata TEXT,
			is_private BOOLEAN DEFAULT FALSE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_session ON memory_entries(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_timestamp ON memory_entries(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_type ON memory_entries(type)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_layer ON memory_entries(layer)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts USING fts5(
			summary, content, tags,
			content=memory_entries,
			content_rowid=rowid,
			tokenize='porter unicode61'
		)`,
		`CREATE TRIGGER IF NOT EXISTS memory_ai AFTER INSERT ON memory_entries BEGIN
			INSERT INTO memory_fts(rowid, summary, content, tags) VALUES (new.rowid, new.summary, new.content, new.tags);
		END`,
		`CREATE TRIGGER IF NOT EXISTS memory_ad AFTER DELETE ON memory_entries BEGIN
			INSERT INTO memory_fts(memory_fts, rowid, summary, content, tags) VALUES('delete', old.rowid, old.summary, old.content, old.tags);
		END`,
		`CREATE TRIGGER IF NOT EXISTS memory_au AFTER UPDATE ON memory_entries BEGIN
			INSERT INTO memory_fts(memory_fts, rowid, summary, content, tags) VALUES('delete', old.rowid, old.summary, old.content, old.tags);
			INSERT INTO memory_fts(rowid, summary, content, tags) VALUES (new.rowid, new.summary, new.content, new.tags);
		END`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

func (s *ProgressiveMemoryStore) Store(entry *MemoryEntry) error {
	tagsJSON, _ := json.Marshal(entry.Tags)
	metaJSON, _ := json.Marshal(entry.Metadata)

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO memory_entries (id, layer, session_id, timestamp, type, summary, content, tags, metadata, is_private)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Layer, entry.SessionID, entry.Timestamp.Format(time.RFC3339Nano),
		entry.Type, entry.Summary, entry.Content, string(tagsJSON), string(metaJSON), entry.IsPrivate,
	)
	return err
}

func (s *ProgressiveMemoryStore) Search(ctx context.Context, query string, layer MemoryLayer, limit int) (*SearchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT e.id, e.layer, e.session_id, e.timestamp, e.type, e.summary, e.content, e.tags, e.metadata, e.is_private
		 FROM memory_entries e
		 JOIN memory_fts f ON e.rowid = f.rowid
		 WHERE memory_fts MATCH ? AND e.layer <= ?
		 ORDER BY rank
		 LIMIT ?`,
		query, layer, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var tagsJSON, metaJSON sql.NullString
		var content sql.NullString
		err := rows.Scan(&e.ID, &e.Layer, &e.SessionID, &e.Timestamp, &e.Type, &e.Summary, &content, &tagsJSON, &metaJSON, &e.IsPrivate)
		if err != nil {
			continue
		}
		if content.Valid {
			e.Content = content.String
		}
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &e.Tags)
		}
		if metaJSON.Valid {
			json.Unmarshal([]byte(metaJSON.String), &e.Metadata)
		}
		entries = append(entries, e)
	}

	tokens := 0
	for _, e := range entries {
		tokens += len(strings.Fields(e.Summary))
		if e.Content != "" {
			tokens += len(strings.Fields(e.Content))
		}
	}

	return &SearchResult{
		Entries:    entries,
		TotalFound: len(entries),
		TokensUsed: tokens,
		Layer:      layer,
	}, nil
}

func (s *ProgressiveMemoryStore) GetTimeline(sessionID string, since time.Time, limit int) ([]MemoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, layer, session_id, timestamp, type, summary, content, tags, metadata, is_private
		 FROM memory_entries
		 WHERE session_id = ? AND timestamp >= ?
		 ORDER BY timestamp ASC
		 LIMIT ?`,
		sessionID, since.Format(time.RFC3339Nano), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var tagsJSON, metaJSON sql.NullString
		var content sql.NullString
		if err := rows.Scan(&e.ID, &e.Layer, &e.SessionID, &e.Timestamp, &e.Type, &e.Summary, &content, &tagsJSON, &metaJSON, &e.IsPrivate); err != nil {
			continue
		}
		if content.Valid {
			e.Content = content.String
		}
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &e.Tags)
		}
		if metaJSON.Valid {
			json.Unmarshal([]byte(metaJSON.String), &e.Metadata)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *ProgressiveMemoryStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM memory_entries WHERE id = ?`, id)
	return err
}

func (s *ProgressiveMemoryStore) Close() error {
	return s.db.Close()
}
