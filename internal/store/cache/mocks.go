package cache

import (
	"context"

	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	args := m.Called(userID)
	return nil, args.Error(1)
}

func (m *MockUserStore) Set(ctx context.Context, user *store.User) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, userID int64) {

}
