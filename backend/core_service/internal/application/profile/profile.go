package profile

import (
	"context"
	"log/slog"

	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id uuid.UUID) (sqlc.GetProfileRow, error)
}

type ProfileService struct {
	repo ProfileRepository
	log  *slog.Logger
}

func NewProfileService(repo ProfileRepository, log *slog.Logger) *ProfileService {
	return &ProfileService{
		repo: repo,
		log:  log,
	}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*sqlc.GetProfileRow, error) {
	res, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
