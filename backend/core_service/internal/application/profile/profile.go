package profile

import (
	"context"
	"log/slog"

	"github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id uuid.UUID) (sqlc.GetProfileRow, error)
	UpdateProfile(ctx context.Context, arg sqlc.UpdateProfileParams) error
	DeleteProfile(ctx context.Context, id uuid.UUID) error
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

func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, updateProfileObject dto.PatchProfileRequest) error {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return err
	}
	arg := sqlc.UpdateProfileParams{
		ID:    userID,
		Name:  profile.Name,
		Email: profile.Email,
	}
	if updateProfileObject.Email != "" {
		arg.Email = updateProfileObject.Email
	}
	if updateProfileObject.Name != "" {
		arg.Name = updateProfileObject.Name
	}
	err = s.repo.UpdateProfile(ctx, arg)
	if err != nil {
		return err
	}
	return nil

}

func (s *ProfileService) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteProfile(ctx, userID)
}
