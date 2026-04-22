package mocks

import (
	"context"
	"database/sql"

	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockQueries struct {
	mock.Mock
}

func (m *MockQueries) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockQueries) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockQueries) GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockQueries) GetUserByInn(ctx context.Context, inn sql.NullString) (sqlc.User, error) {
	args := m.Called(ctx, inn)
	return args.Get(0).(sqlc.User), args.Error(1)
}
