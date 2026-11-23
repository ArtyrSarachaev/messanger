package logic

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"messanger/internal/entity"
)

func TestDialog(t *testing.T) {
	ctx := context.Background()
	users := NewUserLogic(newFakeUserRepo(), fakeTokens{})

	alice, err := users.Register(ctx, entity.Credentials{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	bob, err := users.Register(ctx, entity.Credentials{Username: "bob", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	repo := &fakeMessageRepo{
		dialogResult: []entity.Message{
			{SenderID: bob.ID, ReceiverID: alice.ID, Text: "hi alice", CreatedAt: now},
			{SenderID: alice.ID, ReceiverID: bob.ID, Text: "hi bob", CreatedAt: now.Add(-time.Minute)},
		},
	}
	messages := NewMessageLogic(repo, users)

	t.Run("maps ids to usernames", func(t *testing.T) {
		got, err := messages.Dialog(ctx, alice.ID, "bob", 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d messages, want 2", len(got))
		}
		if got[0].From != "bob" || got[0].To != "alice" {
			t.Errorf("first message: from=%q to=%q", got[0].From, got[0].To)
		}
		if got[1].From != "alice" || got[1].To != "bob" {
			t.Errorf("second message: from=%q to=%q", got[1].From, got[1].To)
		}
		if repo.gotUser1 != alice.ID || repo.gotUser2 != bob.ID {
			t.Errorf("repo called with %q, %q", repo.gotUser1, repo.gotUser2)
		}
	})

	t.Run("normalizes pagination", func(t *testing.T) {
		if _, err := messages.Dialog(ctx, alice.ID, "bob", 0, -5); err != nil {
			t.Fatal(err)
		}
		if repo.gotLimit != DefaultDialogLimit || repo.gotOffset != 0 {
			t.Errorf("limit=%d offset=%d", repo.gotLimit, repo.gotOffset)
		}

		if _, err := messages.Dialog(ctx, alice.ID, "bob", 10_000, 3); err != nil {
			t.Fatal(err)
		}
		if repo.gotLimit != MaxDialogLimit || repo.gotOffset != 3 {
			t.Errorf("limit=%d offset=%d", repo.gotLimit, repo.gotOffset)
		}
	})

	t.Run("unknown companion", func(t *testing.T) {
		_, err := messages.Dialog(ctx, alice.ID, "nobody", 10, 0)
		if !errors.Is(err, entity.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})
}

func TestValidateMessage(t *testing.T) {
	valid := entity.Message{SenderID: "a", ReceiverID: "b", Text: "hello"}

	tests := []struct {
		name    string
		modify  func(*entity.Message)
		wantErr bool
	}{
		{name: "valid", modify: func(*entity.Message) {}},
		{name: "empty text", modify: func(m *entity.Message) { m.Text = "   " }, wantErr: true},
		{name: "no receiver", modify: func(m *entity.Message) { m.ReceiverID = "" }, wantErr: true},
		{name: "too long", modify: func(m *entity.Message) { m.Text = strings.Repeat("я", entity.MaxMessageLength+1) }, wantErr: true},
		{name: "max length", modify: func(m *entity.Message) { m.Text = strings.Repeat("я", entity.MaxMessageLength) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := valid
			tt.modify(&msg)

			err := ValidateMessage(msg)
			if tt.wantErr && !errors.Is(err, entity.ErrInvalidInput) {
				t.Fatalf("want ErrInvalidInput, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSaveRejectsInvalidMessage(t *testing.T) {
	repo := &fakeMessageRepo{}
	messages := NewMessageLogic(repo, NewUserLogic(newFakeUserRepo(), fakeTokens{}))

	err := messages.Save(context.Background(), entity.Message{SenderID: "a", ReceiverID: "b"})
	if !errors.Is(err, entity.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatal("invalid message must not be saved")
	}
}
