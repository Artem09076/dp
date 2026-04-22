package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	log    *slog.Logger
	client *redis.Client
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

func (r *RedisClient) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("user:%s", userID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	var user map[string]interface{}
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *RedisClient) SetUser(ctx context.Context, userID string, userData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("user:%s", userID)
	data, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) GetService(ctx context.Context, serviceID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("service:%s", serviceID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service from cache: %w", err)
	}

	var service map[string]interface{}
	if err := json.Unmarshal(data, &service); err != nil {
		return nil, err
	}
	return service, nil
}

func (r *RedisClient) SetService(ctx context.Context, serviceID string, serviceData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("service:%s", serviceID)
	data, err := json.Marshal(serviceData)
	if err != nil {
		return fmt.Errorf("failed to marshal service data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateService(ctx context.Context, serviceID string) error {
	key := fmt.Sprintf("service:%s", serviceID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) GetUserBookings(ctx context.Context, userID string, role string) ([]byte, error) {
	key := fmt.Sprintf("bookings:user:%s:role:%s", userID, role)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetUserBookings(ctx context.Context, userID string, role string, bookingsData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("bookings:user:%s:role:%s", userID, role)
	data, err := json.Marshal(bookingsData)
	if err != nil {
		return fmt.Errorf("failed to marshal bookings data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateUserBookings(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("bookings:user:%s:role:*", userID)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisClient) GetBooking(ctx context.Context, bookingID string) ([]byte, error) {
	key := fmt.Sprintf("booking:%s", bookingID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get booking from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetBooking(ctx context.Context, bookingID string, bookingData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("booking:%s", bookingID)
	data, err := json.Marshal(bookingData)
	if err != nil {
		return fmt.Errorf("failed to marshal booking data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateBooking(ctx context.Context, bookingID string) error {
	key := fmt.Sprintf("booking:%s", bookingID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
