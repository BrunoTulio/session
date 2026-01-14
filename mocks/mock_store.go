package mocks

import (
	"context"

	"github.com/BrunoTulio/session"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get(ctx context.Context, id string) (session.SessionData, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return session.SessionData{}, args.Error(1)
	}

	return args.Get(0).(session.SessionData), args.Error(1)
}

func (m *MockStore) Set(ctx context.Context, session session.SessionData) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func NewMockStore() session.Store {
	return &MockStore{}
}
