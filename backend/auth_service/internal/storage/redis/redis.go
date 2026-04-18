package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	log    *slog.Logger
	client *redis.Client
}

type SessionData struct {
	UserID       string    `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	DeviceID     string    `json:"device_id"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
}

func NewRedisClient(log *slog.Logger, addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return &RedisClient{log: log, client: client}, nil
}

func (r *RedisClient) SaveSession(ctx context.Context, userID uuid.UUID, deviceID, refreshToken, ipAddress string, ttl time.Duration) error {
	session := SessionData{
		UserID:       userID.String(),
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s:%s", userID.String(), deviceID)

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID.String())
	err = r.client.SAdd(ctx, userSessionsKey, deviceID).Err()
	if err != nil {
		return fmt.Errorf("failed to add to user sessions index: %w", err)
	}

	err = r.client.Expire(ctx, userSessionsKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiry for user sessions index: %w", err)
	}

	return nil
}

func (r *RedisClient) GetSession(ctx context.Context, userID uuid.UUID, deviceID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s:%s", userID.String(), deviceID)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionData
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *RedisClient) UpdateSession(ctx context.Context, userID uuid.UUID, deviceID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s:%s", userID.String(), deviceID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionData
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}

	session.LastUsedAt = time.Now()

	newData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	err = r.client.Set(ctx, key, newData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (r *RedisClient) DeleteSession(ctx context.Context, userID uuid.UUID, deviceID string) error {
	key := fmt.Sprintf("session:%s:%s", userID.String(), deviceID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID.String())
	err = r.client.SRem(ctx, userSessionsKey, deviceID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove from user sessions index: %w", err)
	}

	return nil
}

func (r *RedisClient) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID.String())

	deviceIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, deviceID := range deviceIDs {
		key := fmt.Sprintf("session:%s:%s", userID.String(), deviceID)
		err = r.client.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to delete session for device %s: %w", deviceID, err)
		}
	}

	err = r.client.Del(ctx, userSessionsKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user sessions index: %w", err)
	}

	return nil
}

func (r *RedisClient) GetUserActiveDevices(ctx context.Context, userID uuid.UUID) (int64, error) {
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID.String())
	count, err := r.client.SCard(ctx, userSessionsKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get active devices count: %w", err)
	}
	return count, nil
}

func (r *RedisClient) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	err := r.client.Set(ctx, key, "blacklisted", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

func (r *RedisClient) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	_, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	return true, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
