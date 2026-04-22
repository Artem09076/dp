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

func (r *RedisClient) GetProfile(ctx context.Context, userID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("profile:%s", userID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile from cache: %w", err)
	}

	var profile map[string]interface{}
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *RedisClient) SetProfile(ctx context.Context, userID string, profileData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("profile:%s", userID)
	data, err := json.Marshal(profileData)
	if err != nil {
		return fmt.Errorf("failed to marshal profile data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateProfile(ctx context.Context, userID string) error {
	key := fmt.Sprintf("profile:%s", userID)
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

func (r *RedisClient) GetServicesList(ctx context.Context, performerID string) ([]byte, error) {
	key := fmt.Sprintf("services:performer:%s", performerID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get services list from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetServicesList(ctx context.Context, performerID string, servicesData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("services:performer:%s", performerID)
	data, err := json.Marshal(servicesData)
	if err != nil {
		return fmt.Errorf("failed to marshal services data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateServicesList(ctx context.Context, performerID string) error {
	key := fmt.Sprintf("services:performer:%s", performerID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) SearchServices(ctx context.Context, query string, page, limit int) ([]byte, error) {
	key := fmt.Sprintf("search:services:q:%s:p:%d:l:%d", query, page, limit)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get search results from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetSearchServices(ctx context.Context, query string, page, limit int, results interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("search:services:q:%s:p:%d:l:%d", query, page, limit)
	data, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal search results: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateSearchServices(ctx context.Context) error {
	pattern := "search:services:*"
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisClient) GetDiscount(ctx context.Context, discountID string) ([]byte, error) {
	key := fmt.Sprintf("discount:%s", discountID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get discount from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetDiscount(ctx context.Context, discountID string, discountData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("discount:%s", discountID)
	data, err := json.Marshal(discountData)
	if err != nil {
		return fmt.Errorf("failed to marshal discount data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateDiscount(ctx context.Context, discountID string) error {
	key := fmt.Sprintf("discount:%s", discountID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) GetServiceDiscounts(ctx context.Context, serviceID string) ([]byte, error) {
	key := fmt.Sprintf("discounts:service:%s", serviceID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service discounts from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetServiceDiscounts(ctx context.Context, serviceID string, discountsData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("discounts:service:%s", serviceID)
	data, err := json.Marshal(discountsData)
	if err != nil {
		return fmt.Errorf("failed to marshal discounts data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateServiceDiscounts(ctx context.Context, serviceID string) error {
	key := fmt.Sprintf("discounts:service:%s", serviceID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) GetReview(ctx context.Context, reviewID string) ([]byte, error) {
	key := fmt.Sprintf("review:%s", reviewID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get review from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetReview(ctx context.Context, reviewID string, reviewData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("review:%s", reviewID)
	data, err := json.Marshal(reviewData)
	if err != nil {
		return fmt.Errorf("failed to marshal review data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateReview(ctx context.Context, reviewID string) error {
	key := fmt.Sprintf("review:%s", reviewID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) GetServiceReviews(ctx context.Context, serviceID string, page, limit int) ([]byte, error) {
	key := fmt.Sprintf("reviews:service:%s:p:%d:l:%d", serviceID, page, limit)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service reviews from cache: %w", err)
	}
	return data, nil
}

func (r *RedisClient) SetServiceReviews(ctx context.Context, serviceID string, page, limit int, reviewsData interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("reviews:service:%s:p:%d:l:%d", serviceID, page, limit)
	data, err := json.Marshal(reviewsData)
	if err != nil {
		return fmt.Errorf("failed to marshal reviews data: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) InvalidateServiceReviews(ctx context.Context, serviceID string) error {
	pattern := fmt.Sprintf("reviews:service:%s:*", serviceID)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
