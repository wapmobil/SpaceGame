package auth

import (
	"testing"
	"time"
)

func TestGeneratePlayerID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := GeneratePlayerID()
		if err != nil {
			t.Fatalf("failed to generate player ID: %v", err)
		}
		if ids[id] {
			t.Fatalf("duplicate player ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGeneratePlayerID_Format(t *testing.T) {
	id, err := GeneratePlayerID()
	if err != nil {
		t.Fatalf("failed to generate player ID: %v", err)
	}
	if len(id) != 36 {
		t.Fatalf("expected UUID length 36, got %d", len(id))
	}
}

func TestGenerateAuthToken_Format(t *testing.T) {
	token, err := GenerateAuthToken()
	if err != nil {
		t.Fatalf("failed to generate auth token: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("expected token length 64, got %d", len(token))
	}
}

func TestGenerateAuthToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := GenerateAuthToken()
		if err != nil {
			t.Fatalf("failed to generate auth token: %v", err)
		}
		if tokens[token] {
			t.Fatalf("duplicate auth token generated: %s", token)
		}
		tokens[token] = true
	}
}

func TestNewSession_Expires(t *testing.T) {
	session, err := NewSession("player-1", "token-1")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if session.IsExpired() {
		t.Fatal("new session should not be expired")
	}
	if session.PlayerID != "player-1" {
		t.Fatalf("expected player ID 'player-1', got '%s'", session.PlayerID)
	}
}

func TestSession_Expiration(t *testing.T) {
	session := &Session{
		PlayerID:  "player-1",
		AuthToken: "token-1",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if !session.IsExpired() {
		t.Fatal("session should be expired")
	}
}
