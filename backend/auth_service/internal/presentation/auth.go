package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Artem09076/dp/backend/auth_service/internal/lib/jwt"
	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/auth_service/internal/storage/redis"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log             *slog.Logger
	queries         *sqlc.Queries
	redis           *redis.RedisClient
	tokenAccessTTL  time.Duration
	tokenRefreshTTL time.Duration
	tokenSecret     string
	maxDevices      int
}

func New(log *slog.Logger, queries *sqlc.Queries, redis *redis.RedisClient, tokenAccessTTL, tokenRefreshTTL time.Duration, tokenSecret string, maxDevices int) *Auth {
	return &Auth{
		log:             log,
		queries:         queries,
		redis:           redis,
		tokenAccessTTL:  tokenAccessTTL,
		tokenRefreshTTL: tokenRefreshTTL,
		tokenSecret:     tokenSecret,
		maxDevices:      maxDevices,
	}
}

func (a *Auth) generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *Auth) Register(ctx context.Context, email string, name string, inn string, business_type string, role string, password string, deviceID, ipAddress string) (string, string, error) {
	op := "auth.Auth.Register"
	log := a.log.With("op", op)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to hash password", "error", err)
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}
	var validBusinessTypes sqlc.NullBusinessType
	if business_type != "" {
		validBusinessTypes = sqlc.NullBusinessType{}
		err = validBusinessTypes.Scan(business_type)
		if err != nil {
			a.log.Error("failed to scan business type", "error", err)
			return "", "", fmt.Errorf("failed to scan business type: %w", err)
		}
	} else {
		validBusinessTypes = sqlc.NullBusinessType{Valid: false}
	}

	validUserRole := sqlc.NullUserRole{}
	if err := validUserRole.Scan(role); err != nil {
		a.log.Error("Failed to scan user role", "error", err)
		return "", "", fmt.Errorf("failed to scan user role: %w", err)
	}
	validInn := sql.NullString{
		String: inn,
		Valid:  inn != "",
	}
	var verificationStatus string
	if validUserRole.UserRole == "performer" {
		verificationStatus = "pending"
	} else {
		verificationStatus = "verified"
	}

	user, err := a.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:              email,
		Name:               name,
		Inn:                validInn,
		BusinessType:       validBusinessTypes,
		Role:               validUserRole.UserRole,
		PasswordHash:       passwordHash,
		VerificationStatus: sqlc.VerificationStatus(verificationStatus),
	})

	if err != nil {
		log.Error("failed to create user", "error", err)
		return "", "", fmt.Errorf("failed to create user: %w", err)
	}
	tokenID := a.generateTokenID()

	tokenPair, err := jwt.NewTokenPair(user, deviceID, tokenID, []byte(a.tokenSecret), a.tokenAccessTTL, a.tokenRefreshTTL)
	if err != nil {
		log.Error("failed to create token", "error", err)
		return "", "", fmt.Errorf("failed to create token: %w", err)
	}

	err = a.redis.SaveSession(ctx, user.ID, deviceID, tokenPair.RefreshToken, ipAddress, a.tokenRefreshTTL)
	if err != nil {
		log.Error("failed to save session to Redis", "error", err)
		return "", "", fmt.Errorf("failed to save session: %w", err)
	}

	log.Info("user registered successfully", "user_id", user.ID)

	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}

func (a *Auth) Login1(ctx context.Context, email, password, deviceID, ipAddress string) (string, string, error) {
	op := "auth.Auth.Login1"
	log := a.log.With("op", op)
	user, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("user not found", "email", email)
			return "", "", fmt.Errorf("invalid email or password")
		}
		log.Warn("failed to get user by email", "email", email, "error", err)
		return "", "", fmt.Errorf("failed to get user by email: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Warn("invalid password for user", "email", email)
		return "", "", fmt.Errorf("invalid email")
	}
	return a.generateTokensForUser(ctx, user, deviceID, ipAddress)
}

func (a *Auth) Login2(ctx context.Context, inn, password, deviceID, ipAddress string) (string, string, error) {
	op := "auth.Auth.Login2"
	log := a.log.With("op", op)
	validInn := sql.NullString{
		String: inn,
		Valid:  inn != "",
	}
	if !validInn.Valid {
		log.Warn("invalid inn", "inn", inn)
		return "", "", fmt.Errorf("invalid email or password")
	}
	user, err := a.queries.GetUserByInn(ctx, validInn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("user not found", "inn", inn)
			return "", "", fmt.Errorf("invalid email or password")
		}
		log.Warn("failed to get user by email", "email", inn, "error", err)
		return "", "", fmt.Errorf("failed to get user by email: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Warn("invalid password for user", "inn", inn)
		return "", "", fmt.Errorf("invalid  password")
	}
	return a.generateTokensForUser(ctx, user, deviceID, ipAddress)
}

func (a *Auth) RefreshToken(ctx context.Context, refreshToken, deviceID, ipAddress string) (string, string, error) {
	op := "auth.Auth.RefreshToken"
	log := a.log.With("op", op)

	claims, err := jwt.ParseRefreshToken(refreshToken, []byte(a.tokenSecret))
	if err != nil {
		log.Warn("invalid refresh token", "error", err)
		return "", "", fmt.Errorf("invalid refresh token")
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		log.Warn("invalid user ID in token", "user_id", claims.UserID)
		return "", "", fmt.Errorf("invalid refresh token")
	}

	isBlacklisted, err := a.redis.IsTokenBlacklisted(ctx, refreshToken)
	if err != nil {
		log.Warn("failed to check blacklist", "error", err)
	}
	if isBlacklisted {
		log.Warn("token is blacklisted", "user_id", userID)
		return "", "", fmt.Errorf("token revoked")
	}
	session, err := a.redis.GetSession(ctx, userID, claims.DeviceID)
	if err != nil {
		log.Error("failed to get session from Redis", "error", err)
		return "", "", fmt.Errorf("internal error")
	}
	if session == nil {
		log.Warn("session not found in Redis", "user_id", userID)
		return "", "", fmt.Errorf("invalid refresh token")
	}

	if session.RefreshToken != refreshToken {
		log.Warn("refresh token mismatch")
		return "", "", fmt.Errorf("invalid refresh token")
	}

	user, err := a.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("user not found", "user_id", userID)
			return "", "", fmt.Errorf("user not found")
		}
		log.Error("failed to get user", "user_id", userID, "error", err)
		return "", "", fmt.Errorf("internal error")
	}

	newTokenID := a.generateTokenID()
	newAccessToken, err := jwt.NewAccessToken(user, deviceID, []byte(a.tokenSecret), a.tokenAccessTTL)
	if err != nil {
		log.Error("failed to create new access token", "error", err)
		return "", "", fmt.Errorf("failed to create tokens")
	}

	newRefreshToken, err := jwt.NewRefreshToken(user, deviceID, newTokenID, []byte(a.tokenSecret), a.tokenRefreshTTL)
	if err != nil {
		log.Error("failed to create new refresh token", "error", err)
		return "", "", fmt.Errorf("failed to create tokens")
	}

	err = a.redis.BlacklistToken(ctx, refreshToken, a.tokenRefreshTTL)
	if err != nil {
		log.Warn("failed to blacklist old token", "error", err)
	}

	err = a.redis.SaveSession(ctx, user.ID, deviceID, newRefreshToken, ipAddress, a.tokenRefreshTTL)
	if err != nil {
		log.Error("failed to update session in Redis", "error", err)
		return "", "", fmt.Errorf("failed to update session")
	}

	return newAccessToken, newRefreshToken, nil
}

func (a *Auth) generateTokensForUser(ctx context.Context, user sqlc.User, deviceID, ipAddress string) (string, string, error) {
	log := a.log.With("user_id", user.ID, "deviceID", deviceID)

	activeDevices, err := a.redis.GetUserActiveDevices(ctx, user.ID)
	if err != nil {
		log.Warn("failed to get active devices count", "error", err)
	}

	if activeDevices >= int64(a.maxDevices) {
		log.Warn("max devices limit reached", "active_devices", activeDevices)
		return "", "", fmt.Errorf("maximum devices limit reached (%d)", a.maxDevices)
	}

	if deviceID != "" && deviceID != "unknown" {
		existingSession, _ := a.redis.GetSession(ctx, user.ID, deviceID)
		if existingSession != nil {
			log.Info("existing session found, replacing")
			a.redis.DeleteSession(ctx, user.ID, deviceID)
		}
	}
	tokenID := a.generateTokenID()

	tokenPair, err := jwt.NewTokenPair(user, deviceID, tokenID, []byte(a.tokenSecret), a.tokenAccessTTL, a.tokenRefreshTTL)
	if err != nil {
		log.Error("failed to create token", "error", err)
		return "", "", fmt.Errorf("failed to create token: %w", err)
	}

	err = a.redis.SaveSession(ctx, user.ID, deviceID, tokenPair.RefreshToken, ipAddress, a.tokenRefreshTTL)
	if err != nil {
		log.Error("failed to save session to Redis", "error", err)
		return "", "", fmt.Errorf("failed to save session: %w", err)
	}

	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}

func (a *Auth) Logout(ctx context.Context, userID uuid.UUID, deviceID, accessToken string) error {
	op := "auth.Auth.Logout"
	log := a.log.With("op", op, "user_id", userID, "deviceID", deviceID)

	if accessToken != "" {
		claims, err := jwt.ParseAccessToken(accessToken, []byte(a.tokenSecret))
		if err == nil {
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				a.redis.BlacklistToken(ctx, accessToken, ttl)
			}
		}
	}

	session, err := a.redis.GetSession(ctx, userID, deviceID)
	if err != nil {
		log.Warn("failed to get session", "error", err)
	}

	if session != nil && session.RefreshToken != "" {
		a.redis.BlacklistToken(ctx, session.RefreshToken, a.tokenRefreshTTL)
	}

	err = a.redis.DeleteSession(ctx, userID, deviceID)
	if err != nil {
		log.Error("failed to delete session", "error", err)
		return fmt.Errorf("failed to logout")
	}

	log.Info("user logged out successfully")
	return nil
}

func (a *Auth) ValidateAccessToken(ctx context.Context, tokenString string) (*jwt.AccessClaims, error) {

	isBlacklisted, err := a.redis.IsTokenBlacklisted(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, fmt.Errorf("token is revoked")
	}

	claims, err := jwt.ParseAccessToken(tokenString, []byte(a.tokenSecret))
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
