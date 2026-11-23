package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"messanger/internal/entity"
)

type fakeUserRepo struct {
	mu     sync.Mutex
	nextID int
	byName map[string]entity.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byName: make(map[string]entity.User)}
}

func (r *fakeUserRepo) Create(_ context.Context, username, passwordHash string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byName[username]; ok {
		return entity.User{}, entity.ErrUserAlreadyExists
	}

	r.nextID++
	user := entity.User{
		ID:        fmt.Sprintf("id-%d", r.nextID),
		Username:  username,
		Password:  passwordHash,
		CreatedAt: time.Now(),
	}
	r.byName[username] = user
	return user, nil
}

func (r *fakeUserRepo) ByUsername(_ context.Context, username string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.byName[username]
	if !ok {
		return entity.User{}, entity.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeUserRepo) ByID(_ context.Context, userID string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.byName {
		if user.ID == userID {
			return user, nil
		}
	}
	return entity.User{}, entity.ErrUserNotFound
}

type fakeTokens struct{}

func (fakeTokens) Generate(userID string) (string, error) { return "token-" + userID, nil }

func (fakeTokens) Parse(token string) (string, error) { return token, nil }

type fakeMessageRepo struct {
	saved []entity.Message

	gotUser1, gotUser2  string
	gotLimit, gotOffset int
	dialogResult        []entity.Message
}

func (r *fakeMessageRepo) Save(_ context.Context, message entity.Message) error {
	r.saved = append(r.saved, message)
	return nil
}

func (r *fakeMessageRepo) Dialog(_ context.Context, userID1, userID2 string, limit, offset int) ([]entity.Message, error) {
	r.gotUser1, r.gotUser2 = userID1, userID2
	r.gotLimit, r.gotOffset = limit, offset
	return r.dialogResult, nil
}
