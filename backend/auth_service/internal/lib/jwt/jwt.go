package jwt

import (
	"fmt"
	"time"

	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AccessClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	DeviceID string `json:"device_id"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
	TokenID  string `json:"token_id"`
	jwt.RegisteredClaims
}

func NewTokenPair(user sqlc.User, deviceID, tokenID string, secretKey []byte, accessTTL, refreshTTL time.Duration) (*TokenPair, error) {

	accessToken, err := NewAccessToken(user, deviceID, secretKey, accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := NewRefreshToken(user, deviceID, tokenID, secretKey, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func NewAccessToken(user sqlc.User, deviceID string, secretKey []byte, ttl time.Duration) (string, error) {

	// claims := token.Claims.(jwt.MapClaims)
	// claims["user_id"] = user.ID
	// claims["email"] = user.Email
	// claims["role"] = user.Role
	// claims["type"] = "access"
	// claims["exp"] = time.Now().Add(ttl).Unix()
	// claims["iat"] = time.Now().Unix()

	claims := AccessClaims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Role:     string(user.Role),
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func NewRefreshToken(user sqlc.User, deviceID, tokenID string, secretKey []byte, ttl time.Duration) (string, error) {

	// claims := token.Claims.(jwt.MapClaims)
	// claims["user_id"] = user.ID
	// claims["type"] = "refresh"
	// claims["exp"] = time.Now().Add(ttl).Unix()
	// claims["iat"] = time.Now().Unix()
	claims := RefreshClaims{
		UserID:   user.ID.String(),
		DeviceID: deviceID,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseAccessToken(tokenString string, secret []byte) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func ParseRefreshToken(tokenString string, secret []byte) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
