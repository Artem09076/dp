package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Artem09076/dp/backend/auth_service/internal/lib/jwt"
	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log      *slog.Logger
	queries  *sqlc.Queries
	tokenTTL time.Duration
}

func New(log *slog.Logger, queries *sqlc.Queries, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:      log,
		queries:  queries,
		tokenTTL: tokenTTL,
	}
}

func (a *Auth) Register(ctx context.Context, email string, name string, inn string, business_type string, role string, password string) (string, error) {
	op := "auth.Auth.Register"
	log := a.log.With("op", op)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", "error", err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	validBusinessTypes := sqlc.NullBusinessType{}
	err = validBusinessTypes.Scan(business_type)
	if err != nil {
		log.Error("failed to scan business type", "error", err)
		return "", fmt.Errorf("failed to scan business type: %w", err)
	}
	validUserRole := sqlc.NullUserRole{}
	if err := validUserRole.Scan(role); err != nil {
		log.Error("Failed to scan user role", "error", err)
		return "", fmt.Errorf("failed to scan user role: %w", err)
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
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	token, err := jwt.NewToken(user, []byte("secret"), a.tokenTTL)
	if err != nil {
		log.Error("failed to create token", "error", err)
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

func (a *Auth) Login1(ctx context.Context, email string, password string) (string, error) {
	op := "auth.Auth.Login1"
	log := a.log.With("op", op)
	user, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("user not found", "email", email)
			return "", fmt.Errorf("invalid email or password")
		}
		log.Warn("failed to get user by email", "email", email, "error", err)
		return "", fmt.Errorf("failed to get user by email: %w", err)
	}
	token, err := jwt.NewToken(user, []byte("secret"), a.tokenTTL)
	if err != nil {
		log.Error("failed to create token", "error", err)
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	return token, nil
}

func (a *Auth) Login2(ctx context.Context, inn string, password string) (string, error) {
	op := "auth.Auth.Login2"
	log := a.log.With("op", op)
	validInn := sql.NullString{
		String: inn,
		Valid:  inn != "",
	}
	if !validInn.Valid {
		log.Warn("invalid inn", "inn", inn)
		return "", fmt.Errorf("invalid email or password")
	}
	user, err := a.queries.GetUserByInn(ctx, validInn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("user not found", "inn", inn)
			return "", fmt.Errorf("invalid email or password")
		}
		log.Warn("failed to get user by email", "email", inn, "error", err)
		return "", fmt.Errorf("failed to get user by email: %w", err)
	}
	token, err := jwt.NewToken(user, []byte("secret"), a.tokenTTL)
	if err != nil {
		log.Error("failed to create token", "error", err)
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	return token, nil
}
