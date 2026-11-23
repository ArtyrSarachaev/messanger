package logic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"messanger/internal/entity"
)

func TestRegister(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		creds   entity.Credentials
		prepare func(*testing.T, entity.UserLogic)
		wantErr error
	}{
		{
			name:  "success",
			creds: entity.Credentials{Username: "alice", Password: "secret123"},
		},
		{
			name:  "username is trimmed",
			creds: entity.Credentials{Username: "  bob  ", Password: "secret123"},
		},
		{
			name:    "short username",
			creds:   entity.Credentials{Username: "al", Password: "secret123"},
			wantErr: entity.ErrInvalidInput,
		},
		{
			name:    "username with space",
			creds:   entity.Credentials{Username: "al ice", Password: "secret123"},
			wantErr: entity.ErrInvalidInput,
		},
		{
			name:    "short password",
			creds:   entity.Credentials{Username: "alice", Password: "123"},
			wantErr: entity.ErrInvalidInput,
		},
		{
			name:    "too long password",
			creds:   entity.Credentials{Username: "alice", Password: strings.Repeat("a", 73)},
			wantErr: entity.ErrInvalidInput,
		},
		{
			name:  "duplicate username",
			creds: entity.Credentials{Username: "alice", Password: "secret123"},
			prepare: func(t *testing.T, l entity.UserLogic) {
				if _, err := l.Register(ctx, entity.Credentials{Username: "alice", Password: "other123"}); err != nil {
					t.Fatalf("prepare: %v", err)
				}
			},
			wantErr: entity.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			l := NewUserLogic(repo, fakeTokens{})
			if tt.prepare != nil {
				tt.prepare(t, l)
			}

			user, err := l.Register(ctx, tt.creds)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.Username != strings.TrimSpace(tt.creds.Username) {
				t.Errorf("username = %q, want %q", user.Username, strings.TrimSpace(tt.creds.Username))
			}
			if user.Password == tt.creds.Password {
				t.Error("password must be stored as a hash")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	l := NewUserLogic(repo, fakeTokens{})

	user, err := l.Register(ctx, entity.Credentials{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		token, err := l.Login(ctx, entity.Credentials{Username: "alice", Password: "secret123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "token-"+user.ID {
			t.Errorf("token = %q, want token for %q", token, user.ID)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := l.Login(ctx, entity.Credentials{Username: "alice", Password: "wrong-pass"})
		if !errors.Is(err, entity.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		_, err := l.Login(ctx, entity.Credentials{Username: "nobody", Password: "secret123"})
		if !errors.Is(err, entity.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})
}
