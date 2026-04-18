package jwt

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type TokenBlacklister interface {
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

type JWTValidator struct {
	claims      *jwt.MapClaims
	cache       map[string]interface{}
	secretKey   []byte
	blacklister TokenBlacklister
}

func NewValidator(secretKey string, blacklister TokenBlacklister) *JWTValidator {
	return &JWTValidator{
		claims:      &jwt.MapClaims{},
		cache:       make(map[string]interface{}),
		secretKey:   []byte(secretKey),
		blacklister: blacklister,
	}
}

func (v *JWTValidator) Validate(tokenString string) (*jwt.MapClaims, error) {
	if v.blacklister != nil {
		isBlacklisted, err := v.blacklister.IsTokenBlacklisted(context.Background(), tokenString)
		if err != nil {
			return nil, fmt.Errorf("failed to check blacklist: %w", err)
		}
		if isBlacklisted {
			return nil, errors.New("token is revoked")
		}
	}
	untrustedToken, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %v", err)
	}
	_, ok := untrustedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims structure")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpexted signing method: %v", t.Header["alg"])

		}
		return v.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}
	return nil, errors.New("invalid token")

}
