package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresKV implements PersistentStore using PostgreSQL for key-value storage.
type PostgresKV struct {
	db *sql.DB
}

// NewPostgresKV creates a key-value store backed by PostgreSQL.
func NewPostgresKV(connStr string) (*PostgresKV, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	k := &PostgresKV{db: db}
	if err := k.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return k, nil
}

func (k *PostgresKV) migrate() error {
	_, err := k.db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			agent_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value BYTEA,
			PRIMARY KEY (agent_id, key)
		);
		CREATE INDEX IF NOT EXISTS idx_kv_agent ON kv(agent_id);
	`)
	return err
}

// Get returns the value for an agent and key.
func (k *PostgresKV) Get(ctx context.Context, agentID, key string) ([]byte, error) {
	var val []byte
	err := k.db.QueryRowContext(ctx, `SELECT value FROM kv WHERE agent_id = $1 AND key = $2`, agentID, key).Scan(&val)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return val, err
}

// Set stores a value.
func (k *PostgresKV) Set(ctx context.Context, agentID, key string, value []byte) error {
	_, err := k.db.ExecContext(ctx,
		`INSERT INTO kv (agent_id, key, value) VALUES ($1, $2, $3)
		 ON CONFLICT (agent_id, key) DO UPDATE SET value = EXCLUDED.value`,
		agentID, key, value)
	return err
}

// Delete removes a key.
func (k *PostgresKV) Delete(ctx context.Context, agentID, key string) error {
	_, err := k.db.ExecContext(ctx, `DELETE FROM kv WHERE agent_id = $1 AND key = $2`, agentID, key)
	return err
}

// ListKeys returns all keys for an agent.
func (k *PostgresKV) ListKeys(ctx context.Context, agentID string) ([]string, error) {
	rows, err := k.db.QueryContext(ctx, `SELECT key FROM kv WHERE agent_id = $1`, agentID)
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
func (k *PostgresKV) DeleteAgent(ctx context.Context, agentID string) error {
	_, err := k.db.ExecContext(ctx, `DELETE FROM kv WHERE agent_id = $1`, agentID)
	return err
}

// Close closes the database connection.
func (k *PostgresKV) Close() error {
	return k.db.Close()
}
