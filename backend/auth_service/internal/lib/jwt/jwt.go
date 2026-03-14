package jwt

import (
	"fmt"
	"time"

	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user sqlc.User, secretKey []byte, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["email"] = user.Email
	claims["role"] = user.Role
	claims["verification_status"] = user.VerificationStatus
	claims["exp"] = time.Now().Add(ttl).Unix()

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func DecodeToken(tokenString string, secretKey []byte) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}

}
