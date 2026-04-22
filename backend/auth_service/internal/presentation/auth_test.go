package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/Artem09076/dp/backend/auth_service/internal/lib/jwt"
	"github.com/Artem09076/dp/backend/auth_service/internal/presentation/mocks"
	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/auth_service/internal/storage/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth_Register_Success(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"

	mockQueries.On("CreateUser", mock.Anything, mock.MatchedBy(func(params sqlc.CreateUserParams) bool {
		return params.Email == "test@example.com" &&
			params.Name == "Test User" &&
			params.Role == "user"
	})).Return(sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "user",
	}, nil)

	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Register(
		context.Background(),
		"test@example.com",
		"Test User",
		"",
		"",
		"user",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockQueries.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Register_Performer_Success(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"

	mockQueries.On("CreateUser", mock.Anything, mock.MatchedBy(func(params sqlc.CreateUserParams) bool {
		return params.Email == "performer@example.com" &&
			params.Inn.String == "123456789012" &&
			params.BusinessType.Valid &&
			params.Role == "performer"
	})).Return(sqlc.User{
		ID:    userID,
		Email: "performer@example.com",
		Name:  "Test Performer",
		Role:  "performer",
		Inn:   sql.NullString{String: "123456789012", Valid: true},
	}, nil)

	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Register(
		context.Background(),
		"performer@example.com",
		"Test Performer",
		"123456789012",
		"LLC",
		"performer",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockQueries.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Register_DuplicateEmail(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	mockQueries.On("CreateUser", mock.Anything, mock.Anything).Return(sqlc.User{}, errors.New("duplicate key value violates unique constraint"))

	accessToken, refreshToken, err := auth.Register(
		context.Background(),
		"existing@example.com",
		"Test User",
		"",
		"",
		"user",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockQueries.AssertExpectations(t)
}

func TestAuth_Register_DatabaseError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	mockQueries.On("CreateUser", mock.Anything, mock.Anything).Return(sqlc.User{}, errors.New("database connection error"))

	accessToken, refreshToken, err := auth.Register(
		context.Background(),
		"test@example.com",
		"Test User",
		"",
		"",
		"user",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockQueries.AssertExpectations(t)
}

func TestAuth_Login1_Success(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: hashedPassword,
	}, nil)

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(0), nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, nil)
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockQueries.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Login1_WithExistingSession(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: hashedPassword,
	}, nil)

	existingSession := &redis.SessionData{
		UserID:       userID.String(),
		RefreshToken: "old-refresh-token",
		DeviceID:     "device123",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(1), nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(existingSession, nil)
	mockRedis.On("DeleteSession", mock.Anything, userID, "device123").Return(nil)
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Login1_InvalidPassword(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct123"), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
	}, nil)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		"wrong123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuth_Login1_UserNotFound(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	mockQueries.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(sqlc.User{}, sql.ErrNoRows)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"notfound@example.com",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email or password")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuth_Login1_DatabaseError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{}, errors.New("database connection error"))

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user by email")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuth_Login2_Success(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	inn := sql.NullString{String: "123456789012", Valid: true}

	mockQueries.On("GetUserByInn", mock.Anything, inn).Return(sqlc.User{
		ID:           userID,
		Inn:          inn,
		Role:         "performer",
		PasswordHash: hashedPassword,
	}, nil)

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(0), nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, nil)
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Login2(
		context.Background(),
		"123456789012",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockQueries.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Login2_UserNotFound(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	inn := sql.NullString{String: "123456789012", Valid: true}
	mockQueries.On("GetUserByInn", mock.Anything, inn).Return(sqlc.User{}, sql.ErrNoRows)

	accessToken, refreshToken, err := auth.Login2(
		context.Background(),
		"123456789012",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email or password")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuth_MaxDevicesLimit(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 2)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
	}, nil)

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(2), nil)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum devices limit reached")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuth_Logout_Success(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()

	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}
	accessToken, err := jwt.NewAccessToken(user, "device123", []byte("test-secret"), time.Hour)
	require.NoError(t, err)

	session := &redis.SessionData{
		UserID:       userID.String(),
		RefreshToken: "refresh-token",
		DeviceID:     "device123",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(session, nil)
	mockRedis.On("BlacklistToken", mock.Anything, mock.MatchedBy(func(token string) bool {
		return token == accessToken
	}), mock.Anything).Return(nil)
	mockRedis.On("BlacklistToken", mock.Anything, "refresh-token", 168*time.Hour).Return(nil)
	mockRedis.On("DeleteSession", mock.Anything, userID, "device123").Return(nil)

	err = auth.Logout(context.Background(), userID, "device123", accessToken)

	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Logout_NoSession(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()

	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}
	accessToken, err := jwt.NewAccessToken(user, "device123", []byte("test-secret"), time.Hour)
	require.NoError(t, err)

	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, nil)
	mockRedis.On("BlacklistToken", mock.Anything, accessToken, mock.Anything).Return(nil)
	mockRedis.On("DeleteSession", mock.Anything, userID, "device123").Return(nil)

	err = auth.Logout(context.Background(), userID, "device123", accessToken)

	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Logout_WithInvalidToken(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	invalidToken := "invalid-token-format"

	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(&redis.SessionData{
		UserID:       userID.String(),
		RefreshToken: "refresh-token",
		DeviceID:     "device123",
	}, nil)
	mockRedis.On("BlacklistToken", mock.Anything, "refresh-token", 168*time.Hour).Return(nil)
	mockRedis.On("DeleteSession", mock.Anything, userID, "device123").Return(nil)

	err := auth.Logout(context.Background(), userID, "device123", invalidToken)

	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestAuth_RedisSaveError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: hashedPassword,
	}, nil)

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(0), nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, nil)
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(errors.New("redis connection error"))

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save session")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockRedis.AssertExpectations(t)
}

func TestAuth_RefreshToken_Success(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}

	refreshToken, err := jwt.NewRefreshToken(user, "device123", "token123", []byte("test-secret"), 168*time.Hour)
	require.NoError(t, err)

	session := &redis.SessionData{
		UserID:       userID.String(),
		RefreshToken: refreshToken,
		DeviceID:     "device123",
		IPAddress:    "127.0.0.1",
	}

	mockRedis.On("IsTokenBlacklisted", mock.Anything, refreshToken).Return(false, nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(session, nil)
	mockQueries.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockRedis.On("BlacklistToken", mock.Anything, refreshToken, 168*time.Hour).Return(nil)
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	newAccessToken, newRefreshToken, err := auth.RefreshToken(
		context.Background(),
		refreshToken,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	mockRedis.AssertExpectations(t)
	mockQueries.AssertExpectations(t)
}

func TestAuth_RefreshToken_Blacklisted(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}

	refreshToken, err := jwt.NewRefreshToken(user, "device123", "token123", []byte("test-secret"), 168*time.Hour)
	require.NoError(t, err)

	mockRedis.On("IsTokenBlacklisted", mock.Anything, refreshToken).Return(true, nil)

	newAccessToken, newRefreshToken, err := auth.RefreshToken(
		context.Background(),
		refreshToken,
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token revoked")
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	mockRedis.AssertExpectations(t)
}

func TestAuth_RefreshToken_InvalidToken(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	invalidToken := "invalid-token-format"

	newAccessToken, newRefreshToken, err := auth.RefreshToken(
		context.Background(),
		invalidToken,
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
}

func TestAuth_RefreshToken_SessionNotFound(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}

	refreshToken, err := jwt.NewRefreshToken(user, "device123", "token123", []byte("test-secret"), 168*time.Hour)
	require.NoError(t, err)

	mockRedis.On("IsTokenBlacklisted", mock.Anything, refreshToken).Return(false, nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, nil)

	newAccessToken, newRefreshToken, err := auth.RefreshToken(
		context.Background(),
		refreshToken,
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
}

func TestAuth_RefreshToken_TokenMismatch(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}

	refreshToken, err := jwt.NewRefreshToken(user, "device123", "token123", []byte("test-secret"), 168*time.Hour)
	require.NoError(t, err)

	session := &redis.SessionData{
		UserID:       userID.String(),
		RefreshToken: "different-token",
		DeviceID:     "device123",
	}

	mockRedis.On("IsTokenBlacklisted", mock.Anything, refreshToken).Return(false, nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(session, nil)

	newAccessToken, newRefreshToken, err := auth.RefreshToken(
		context.Background(),
		refreshToken,
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
}

func TestAuth_RefreshToken_UserNotFound(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	user := sqlc.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  "user",
	}

	refreshToken, err := jwt.NewRefreshToken(user, "device123", "token123", []byte("test-secret"), 168*time.Hour)
	require.NoError(t, err)

	session := &redis.SessionData{
		UserID:       userID.String(),
		RefreshToken: refreshToken,
		DeviceID:     "device123",
	}

	mockRedis.On("IsTokenBlacklisted", mock.Anything, refreshToken).Return(false, nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(session, nil)
	mockQueries.On("GetUserByID", mock.Anything, userID).Return(sqlc.User{}, sql.ErrNoRows)

	newAccessToken, newRefreshToken, err := auth.RefreshToken(
		context.Background(),
		refreshToken,
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
}

func TestAuth_ValidateAccessToken_RedisError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	token := "some-token"

	mockRedis.On("IsTokenBlacklisted", mock.Anything, token).Return(false, errors.New("redis connection error"))

	claims, err := auth.ValidateAccessToken(context.Background(), token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check blacklist")
	assert.Nil(t, claims)
}

func TestAuth_ValidateAccessToken_InvalidToken(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	invalidToken := "invalid-token-format"

	mockRedis.On("IsTokenBlacklisted", mock.Anything, invalidToken).Return(false, nil)

	claims, err := auth.ValidateAccessToken(context.Background(), invalidToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuth_Login1_GetUserActiveDevicesError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: hashedPassword,
	}, nil)

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(0), errors.New("redis error"))
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, nil)
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Login1_GetSessionError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(sqlc.User{
		ID:           userID,
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: hashedPassword,
	}, nil)

	mockRedis.On("GetUserActiveDevices", mock.Anything, userID).Return(int64(0), nil)
	mockRedis.On("GetSession", mock.Anything, userID, "device123").Return(nil, errors.New("redis connection error"))
	mockRedis.On("SaveSession", mock.Anything, userID, "device123", mock.Anything, "127.0.0.1", 168*time.Hour).Return(nil)

	accessToken, refreshToken, err := auth.Login1(
		context.Background(),
		"test@example.com",
		password,
		"device123",
		"127.0.0.1",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockRedis.AssertExpectations(t)
}

func TestAuth_Register_InvalidRole(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	mockQueries.On("CreateUser", mock.Anything, mock.MatchedBy(func(params sqlc.CreateUserParams) bool {
		return params.Role == "invalid_role"
	})).Return(sqlc.User{}, errors.New("invalid role"))

	accessToken, refreshToken, err := auth.Register(
		context.Background(),
		"test@example.com",
		"Test User",
		"",
		"",
		"invalid_role",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockQueries.AssertExpectations(t)
}

func TestAuth_Register_InvalidBusinessType(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	mockQueries.On("CreateUser", mock.Anything, mock.MatchedBy(func(params sqlc.CreateUserParams) bool {
		return params.BusinessType.BusinessType == "INVALID_TYPE"
	})).Return(sqlc.User{}, errors.New("invalid business type"))

	accessToken, refreshToken, err := auth.Register(
		context.Background(),
		"performer@example.com",
		"Test Performer",
		"123456789012",
		"INVALID_TYPE",
		"performer",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockQueries.AssertExpectations(t)
}
func TestAuth_Login2_InvalidPassword(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct123"), bcrypt.DefaultCost)
	inn := sql.NullString{String: "123456789012", Valid: true}

	mockQueries.On("GetUserByInn", mock.Anything, inn).Return(sqlc.User{
		ID:           uuid.New(),
		Inn:          inn,
		PasswordHash: hashedPassword,
	}, nil)

	accessToken, refreshToken, err := auth.Login2(
		context.Background(),
		"123456789012",
		"wrong123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid  password")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuth_Login2_DatabaseError(t *testing.T) {
	mockQueries := new(mocks.MockQueries)
	mockRedis := new(mocks.MockRedisClient)
	logger := slog.Default()

	auth := New(logger, mockQueries, mockRedis, time.Hour, 168*time.Hour, "test-secret", 5)

	inn := sql.NullString{String: "123456789012", Valid: true}
	mockQueries.On("GetUserByInn", mock.Anything, inn).Return(sqlc.User{}, errors.New("database connection error"))

	accessToken, refreshToken, err := auth.Login2(
		context.Background(),
		"123456789012",
		"password123",
		"device123",
		"127.0.0.1",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user by email")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}
