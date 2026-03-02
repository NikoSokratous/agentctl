package store

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

// SQLiteKV implements PersistentStore using SQLite for key-value storage.
// This is a minimal implementation for the memory package.
type SQLiteKV struct {
	db *sql.DB
}

// NewSQLiteKV creates a key-value store backed by SQLite.
func NewSQLiteKV(path string) (*SQLiteKV, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	k := &SQLiteKV{db: db}
	if err := k.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return k, nil
}

func (k *SQLiteKV) migrate() error {
	_, err := k.db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			agent_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value BLOB,
			PRIMARY KEY (agent_id, key)
		);
		CREATE INDEX IF NOT EXISTS idx_kv_agent ON kv(agent_id);
	`)
	return err
}

// Get returns the value for an agent and key.
func (k *SQLiteKV) Get(ctx context.Context, agentID, key string) ([]byte, error) {
	var val []byte
	err := k.db.QueryRowContext(ctx, `SELECT value FROM kv WHERE agent_id = ? AND key = ?`, agentID, key).Scan(&val)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return val, err
}

// Set stores a value.
func (k *SQLiteKV) Set(ctx context.Context, agentID, key string, value []byte) error {
	_, err := k.db.ExecContext(ctx, `INSERT OR REPLACE INTO kv (agent_id, key, value) VALUES (?, ?, ?)`, agentID, key, value)
	return err
}

// Delete removes a key.
func (k *SQLiteKV) Delete(ctx context.Context, agentID, key string) error {
	_, err := k.db.ExecContext(ctx, `DELETE FROM kv WHERE agent_id = ? AND key = ?`, agentID, key)
	return err
}

// ListKeys returns all keys for an agent.
func (k *SQLiteKV) ListKeys(ctx context.Context, agentID string) ([]string, error) {
	rows, err := k.db.QueryContext(ctx, `SELECT key FROM kv WHERE agent_id = ?`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// DeleteAgent removes all keys for an agent (GDPR delete).
func (k *SQLiteKV) DeleteAgent(ctx context.Context, agentID string) error {
	_, err := k.db.ExecContext(ctx, `DELETE FROM kv WHERE agent_id = ?`, agentID)
	return err
}

// Close closes the database.
func (k *SQLiteKV) Close() error {
	return k.db.Close()
}
