package auth

import (
	"testing"
	"time"
)

func TestJWTManager(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)

	token, err := m.Generate("user-1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		userID, err := m.Parse(token)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if userID != "user-1" {
			t.Errorf("userID = %q, want user-1", userID)
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		other := NewJWTManager("other-secret", time.Hour)
		if _, err := other.Parse(token); err != ErrInvalidToken {
			t.Fatalf("want ErrInvalidToken, got %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expired := NewJWTManager("test-secret", time.Hour)
		expired.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
		if _, err := expired.Parse(token); err != ErrInvalidToken {
			t.Fatalf("want ErrInvalidToken, got %v", err)
		}
	})

	t.Run("garbage", func(t *testing.T) {
		if _, err := m.Parse("not-a-token"); err != ErrInvalidToken {
			t.Fatalf("want ErrInvalidToken, got %v", err)
		}
	})
}
