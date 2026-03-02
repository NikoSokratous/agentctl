package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"
)

// AuditLogger handles encrypted audit log storage
type AuditLogger struct {
	db             *sql.DB
	keyManager     KeyManager
	currentKey     *EncryptionKey
	mu             sync.RWMutex
	rotationPeriod time.Duration
	stopChan       chan struct{}
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID            string
	TenantID      string
	UserID        string
	Action        string
	Resource      string
	ResourceID    string
	Status        string
	Details       string // Encrypted
	IPAddress     string
	UserAgent     string
	Timestamp     time.Time
	EncryptionKey string // Key version used
}

// EncryptionKey represents a key for encryption
type EncryptionKey struct {
	ID        string
	Version   int
	Key       []byte
	CreatedAt time.Time
	ExpiresAt time.Time
	Active    bool
}

// KeyManager handles encryption key management and rotation
type KeyManager interface {
	GetCurrentKey(ctx context.Context) (*EncryptionKey, error)
	GetKey(ctx context.Context, version int) (*EncryptionKey, error)
	RotateKey(ctx context.Context) (*EncryptionKey, error)
	RevokeKey(ctx context.Context, version int) error
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sql.DB, keyManager KeyManager) *AuditLogger {
	return &AuditLogger{
		db:             db,
		keyManager:     keyManager,
		rotationPeriod: 90 * 24 * time.Hour, // 90 days
		stopChan:       make(chan struct{}),
	}
}

// Start starts the audit logger background tasks
func (al *AuditLogger) Start(ctx context.Context) error {
	// Get current key
	key, err := al.keyManager.GetCurrentKey(ctx)
	if err != nil {
		return fmt.Errorf("get current key: %w", err)
	}

	al.mu.Lock()
	al.currentKey = key
	al.mu.Unlock()

	// Start key rotation loop
	go al.keyRotationLoop(ctx)

	return nil
}

// Stop stops the audit logger
func (al *AuditLogger) Stop() {
	close(al.stopChan)
}

// Log logs an audit entry
func (al *AuditLogger) Log(ctx context.Context, entry *AuditLogEntry) error {
	// Get current key
	al.mu.RLock()
	key := al.currentKey
	al.mu.RUnlock()

	if key == nil {
		return fmt.Errorf("no encryption key available")
	}

	// Encrypt details
	encryptedDetails, err := al.encrypt(entry.Details, key.Key)
	if err != nil {
		return fmt.Errorf("encrypt details: %w", err)
	}

	// Store in database
	query := `
		INSERT INTO audit_logs_encrypted (
			id, tenant_id, user_id, action, resource, resource_id,
			status, details_encrypted, ip_address, user_agent,
			timestamp, encryption_key_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = al.db.ExecContext(ctx, query,
		entry.ID,
		entry.TenantID,
		entry.UserID,
		entry.Action,
		entry.Resource,
		entry.ResourceID,
		entry.Status,
		encryptedDetails,
		entry.IPAddress,
		entry.UserAgent,
		entry.Timestamp,
		key.Version,
	)

	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

// Query queries audit logs with decryption
func (al *AuditLogger) Query(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*AuditLogEntry, error) {
	query := `
		SELECT id, tenant_id, user_id, action, resource, resource_id,
		       status, details_encrypted, ip_address, user_agent,
		       timestamp, encryption_key_version
		FROM audit_logs_encrypted
		WHERE tenant_id = ?
	`

	args := []interface{}{tenantID}

	// Add filters
	if action, ok := filters["action"].(string); ok && action != "" {
		query += " AND action = ?"
		args = append(args, action)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query += " AND resource = ?"
		args = append(args, resource)
	}

	query += " ORDER BY timestamp DESC LIMIT 100"

	rows, err := al.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var entries []*AuditLogEntry
	for rows.Next() {
		entry := &AuditLogEntry{}
		var encryptedDetails string
		var keyVersion int

		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.UserID,
			&entry.Action, &entry.Resource, &entry.ResourceID,
			&entry.Status, &encryptedDetails, &entry.IPAddress,
			&entry.UserAgent, &entry.Timestamp, &keyVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		// Get key for decryption
		key, err := al.keyManager.GetKey(ctx, keyVersion)
		if err != nil {
			return nil, fmt.Errorf("get key version %d: %w", keyVersion, err)
		}

		// Decrypt details
		details, err := al.decrypt(encryptedDetails, key.Key)
		if err != nil {
			return nil, fmt.Errorf("decrypt details: %w", err)
		}

		entry.Details = details
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// encrypt encrypts data using AES-GCM
func (al *AuditLogger) encrypt(plaintext string, key []byte) (string, error) {
	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-GCM
func (al *AuditLogger) decrypt(ciphertext string, key []byte) (string, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// keyRotationLoop periodically rotates encryption keys
func (al *AuditLogger) keyRotationLoop(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-al.stopChan:
			return
		case <-ticker.C:
			al.checkAndRotateKey(ctx)
		}
	}
}

// checkAndRotateKey checks if key rotation is needed
func (al *AuditLogger) checkAndRotateKey(ctx context.Context) {
	al.mu.RLock()
	currentKey := al.currentKey
	al.mu.RUnlock()

	if currentKey == nil {
		return
	}

	// Check if key is expired or nearing expiration
	if time.Until(currentKey.ExpiresAt) < 7*24*time.Hour {
		// Rotate key
		newKey, err := al.keyManager.RotateKey(ctx)
		if err != nil {
			fmt.Printf("Failed to rotate key: %v\n", err)
			return
		}

		al.mu.Lock()
		al.currentKey = newKey
		al.mu.Unlock()

		fmt.Printf("Rotated encryption key to version %d\n", newKey.Version)
	}
}

// LocalKeyManager implements KeyManager with local storage
type LocalKeyManager struct {
	db         *sql.DB
	keys       map[int]*EncryptionKey
	currentVer int
	mu         sync.RWMutex
}

// NewLocalKeyManager creates a new local key manager
func NewLocalKeyManager(db *sql.DB) *LocalKeyManager {
	return &LocalKeyManager{
		db:   db,
		keys: make(map[int]*EncryptionKey),
	}
}

// GetCurrentKey returns the current active key
func (km *LocalKeyManager) GetCurrentKey(ctx context.Context) (*EncryptionKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.currentVer == 0 {
		// Generate first key
		return km.generateKey(ctx)
	}

	key, ok := km.keys[km.currentVer]
	if !ok {
		return nil, fmt.Errorf("current key not found")
	}

	return key, nil
}

// GetKey returns a key by version
func (km *LocalKeyManager) GetKey(ctx context.Context, version int) (*EncryptionKey, error) {
	km.mu.RLock()
	key, ok := km.keys[version]
	km.mu.RUnlock()

	if ok {
		return key, nil
	}

	// Load from database
	return km.loadKey(ctx, version)
}

// RotateKey creates a new encryption key
func (km *LocalKeyManager) RotateKey(ctx context.Context) (*EncryptionKey, error) {
	return km.generateKey(ctx)
}

// RevokeKey revokes a key version
func (km *LocalKeyManager) RevokeKey(ctx context.Context, version int) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	if key, ok := km.keys[version]; ok {
		key.Active = false
	}

	// Update in database
	_, err := km.db.ExecContext(ctx, `
		UPDATE encryption_keys SET active = 0 WHERE version = ?
	`, version)

	return err
}

// generateKey generates a new encryption key
func (km *LocalKeyManager) generateKey(ctx context.Context) (*EncryptionKey, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Generate random key
	keyBytes := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	km.currentVer++
	key := &EncryptionKey{
		ID:        fmt.Sprintf("key-%d", km.currentVer),
		Version:   km.currentVer,
		Key:       keyBytes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year
		Active:    true,
	}

	// Store in database
	keyHash := sha256.Sum256(keyBytes)
	keyHashB64 := base64.StdEncoding.EncodeToString(keyHash[:])

	_, err := km.db.ExecContext(ctx, `
		INSERT INTO encryption_keys (
			id, version, key_hash, created_at, expires_at, active
		) VALUES (?, ?, ?, ?, ?, ?)
	`, key.ID, key.Version, keyHashB64, key.CreatedAt, key.ExpiresAt, 1)

	if err != nil {
		return nil, fmt.Errorf("store key: %w", err)
	}

	km.keys[key.Version] = key
	return key, nil
}

// loadKey loads a key from database
func (km *LocalKeyManager) loadKey(ctx context.Context, version int) (*EncryptionKey, error) {
	// In production, this would load the actual key from KMS
	// For now, return error as keys should be in memory
	return nil, fmt.Errorf("key version %d not in memory", version)
}

// KMSKeyManager implements KeyManager with AWS KMS, GCP KMS, or Vault
type KMSKeyManager struct {
	provider string // aws, gcp, vault
	client   interface{}
	mu       sync.RWMutex
}

// NewKMSKeyManager creates a KMS-backed key manager
func NewKMSKeyManager(provider string) *KMSKeyManager {
	return &KMSKeyManager{
		provider: provider,
	}
}

// GetCurrentKey returns the current key from KMS
func (km *KMSKeyManager) GetCurrentKey(ctx context.Context) (*EncryptionKey, error) {
	// In production, integrate with actual KMS
	// AWS KMS: kms.GenerateDataKey
	// GCP KMS: cloudkms.Encrypt
	// Vault: vault.Transit.Encrypt
	return nil, fmt.Errorf("KMS integration not implemented")
}

// GetKey returns a key by version from KMS
func (km *KMSKeyManager) GetKey(ctx context.Context, version int) (*EncryptionKey, error) {
	return nil, fmt.Errorf("KMS integration not implemented")
}

// RotateKey rotates the KMS key
func (km *KMSKeyManager) RotateKey(ctx context.Context) (*EncryptionKey, error) {
	return nil, fmt.Errorf("KMS integration not implemented")
}

// RevokeKey revokes a KMS key
func (km *KMSKeyManager) RevokeKey(ctx context.Context, version int) error {
	return fmt.Errorf("KMS integration not implemented")
}
