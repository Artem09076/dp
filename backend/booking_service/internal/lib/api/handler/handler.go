package handlerlib

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func GetUserIDFromClaims(ctx context.Context) (uuid.UUID, error) {
	userIDstr, ok := ctx.Value("user_id").(string)
	if !ok || userIDstr == "" {
		return uuid.Nil, errors.New("invalid authentication claims")
	}
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user_id")
	}
	return userID, nil
}
