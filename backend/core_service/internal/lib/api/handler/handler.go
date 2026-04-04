package handlerlib

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func GetUserIDFromClaims(ctx context.Context) (uuid.UUID, error) {
	// claims, ok := ctx.Value("claims").(*jwt.MapClaims)
	// if !ok || claims == nil {
	// 	return uuid.Nil, errors.New("invalid authentication claims")
	// }
	userID, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user_id")
	}
	return userID, nil
}

func GetPaginationParams(ctx context.Context) (page int, limit int) {
	page = 1
	limit = 10

	if ctx.Value("page") != nil {
		page = ctx.Value("page").(int)
	}
	if ctx.Value("limit") != nil {
		limit = ctx.Value("limit").(int)
	}
	return page, limit
}
