package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	claims    *jwt.MapClaims
	cache     map[string]interface{}
	secretKey []byte
}

func NewValidator(secretKey string) *JWTValidator {
	return &JWTValidator{
		claims:    &jwt.MapClaims{},
		cache:     make(map[string]interface{}),
		secretKey: []byte(secretKey),
	}
}

func (v *JWTValidator) Validate(tokenString string) (*jwt.MapClaims, error) {
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
