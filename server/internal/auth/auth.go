package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GeneratePlayerID creates a UUID v7 for a unique player identifier.
// UUID v7 provides time-sortable UUIDs based on RFC 9562.
func GeneratePlayerID() (string, error) {
	id := uuid.Must(uuid.NewV7()).String()
	return id, nil
}

// GenerateAuthToken creates a random 32-byte hex token for session authentication.
func GenerateAuthToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate auth token: %w", err)
	}
	return fmt.Sprintf("%x", bytes), nil
}

// Session represents an active player session.
type Session struct {
	PlayerID  string
	AuthToken string
	ExpiresAt time.Time
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// NewSession creates a new session valid for 24 hours.
func NewSession(playerID, authToken string) (*Session, error) {
	token, err := GenerateAuthToken()
	if err != nil {
		return nil, err
	}
	return &Session{
		PlayerID:  playerID,
		AuthToken: token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}
