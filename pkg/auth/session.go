package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

// Session represents a user session.
type Session struct {
	ID           string        `json:"id"`
	UserInfo     *UserInfo     `json:"user_info"`
	Token        *oauth2.Token `json:"-"` // Not exposed in JSON
	CreatedAt    time.Time     `json:"created_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
	LastActivity time.Time     `json:"last_activity"`
	IPAddress    string        `json:"ip_address,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
}

// Valid checks if session is valid.
func (s *Session) Valid() bool {
	return time.Now().Before(s.ExpiresAt)
}

// SessionManager manages user sessions.
type SessionManager struct {
	db              *sql.DB
	sessionDuration time.Duration
	cleanupInterval time.Duration
}

// NewSessionManager creates a new session manager.
func NewSessionManager(db *sql.DB) *SessionManager {
	sm := &SessionManager{
		db:              db,
		sessionDuration: 24 * time.Hour, // Default 24 hours
		cleanupInterval: 1 * time.Hour,  // Cleanup every hour
	}

	// Start background cleanup
	go sm.periodicCleanup()

	return sm
}

// CreateSession creates a new session.
func (sm *SessionManager) CreateSession(ctx context.Context, userInfo *UserInfo, token *oauth2.Token) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("generate session ID: %w", err)
	}

	session := &Session{
		ID:           sessionID,
		UserInfo:     userInfo,
		Token:        token,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(sm.sessionDuration),
		LastActivity: time.Now(),
	}

	// Store in database
	if err := sm.storeSession(ctx, session); err != nil {
		return nil, fmt.Errorf("store session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT user_id, user_email, user_name, created_at, expires_at, last_activity
		FROM sessions
		WHERE id = ? AND expires_at > ?
	`

	var session Session
	session.ID = sessionID

	var userID, userEmail, userName string

	err := sm.db.QueryRowContext(ctx, query, sessionID, time.Now()).Scan(
		&userID,
		&userEmail,
		&userName,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
	)
	if err != nil {
		return nil, err
	}

	session.UserInfo = &UserInfo{
		ID:    userID,
		Email: userEmail,
		Name:  userName,
	}

	// Update last activity
	go sm.updateLastActivity(context.Background(), sessionID)

	return &session, nil
}

// RevokeSession revokes a session.
func (sm *SessionManager) RevokeSession(ctx context.Context, sessionID string) error {
	query := "DELETE FROM sessions WHERE id = ?"
	_, err := sm.db.ExecContext(ctx, query, sessionID)
	return err
}

// RevokeUserSessions revokes all sessions for a user.
func (sm *SessionManager) RevokeUserSessions(ctx context.Context, userID string) error {
	query := "DELETE FROM sessions WHERE user_id = ?"
	_, err := sm.db.ExecContext(ctx, query, userID)
	return err
}

// ListUserSessions lists all active sessions for a user.
func (sm *SessionManager) ListUserSessions(ctx context.Context, userID string) ([]Session, error) {
	query := `
		SELECT id, created_at, expires_at, last_activity, ip_address, user_agent
		FROM sessions
		WHERE user_id = ? AND expires_at > ?
		ORDER BY last_activity DESC
	`

	rows, err := sm.db.QueryContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]Session, 0)

	for rows.Next() {
		var s Session
		err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.ExpiresAt,
			&s.LastActivity,
			&s.IPAddress,
			&s.UserAgent,
		)
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// CleanupExpired removes expired sessions.
func (sm *SessionManager) CleanupExpired(ctx context.Context) (int64, error) {
	query := "DELETE FROM sessions WHERE expires_at <= ?"
	result, err := sm.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// storeSession stores a session in the database.
func (sm *SessionManager) storeSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, user_email, user_name, created_at, expires_at, last_activity
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := sm.db.ExecContext(ctx, query,
		session.ID,
		session.UserInfo.ID,
		session.UserInfo.Email,
		session.UserInfo.Name,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastActivity,
	)

	return err
}

// updateLastActivity updates the last activity timestamp.
func (sm *SessionManager) updateLastActivity(ctx context.Context, sessionID string) {
	query := "UPDATE sessions SET last_activity = ? WHERE id = ?"
	sm.db.ExecContext(ctx, query, time.Now(), sessionID)
}

// periodicCleanup runs periodic cleanup of expired sessions.
func (sm *SessionManager) periodicCleanup() {
	ticker := time.NewTicker(sm.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		sm.CleanupExpired(ctx)
		cancel()
	}
}

// generateSessionID generates a secure random session ID.
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SetSessionDuration sets the session duration.
func (sm *SessionManager) SetSessionDuration(duration time.Duration) {
	sm.sessionDuration = duration
}
