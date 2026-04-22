package mocks

import (
	"context"
	"time"

	"github.com/Artem09076/dp/backend/auth_service/internal/storage/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient - мок для Redis клиента
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) SaveSession(ctx context.Context, userID uuid.UUID, deviceID, refreshToken, ipAddress string, ttl time.Duration) error {
	args := m.Called(ctx, userID, deviceID, refreshToken, ipAddress, ttl)
	return args.Error(0)
}

func (m *MockRedisClient) GetSession(ctx context.Context, userID uuid.UUID, deviceID string) (*redis.SessionData, error) {
	args := m.Called(ctx, userID, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*redis.SessionData), args.Error(1)
}

func (m *MockRedisClient) UpdateSession(ctx context.Context, userID uuid.UUID, deviceID string, ttl time.Duration) error {
	args := m.Called(ctx, userID, deviceID, ttl)
	return args.Error(0)
}

func (m *MockRedisClient) DeleteSession(ctx context.Context, userID uuid.UUID, deviceID string) error {
	args := m.Called(ctx, userID, deviceID)
	return args.Error(0)
}

func (m *MockRedisClient) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRedisClient) GetUserActiveDevices(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	args := m.Called(ctx, token, ttl)
	return args.Error(0)
}

func (m *MockRedisClient) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
