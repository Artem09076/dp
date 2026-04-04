package profile

import (
	"context"
	"log/slog"

	"github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/rabbit"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id uuid.UUID) (sqlc.GetProfileRow, error)
	UpdateProfile(ctx context.Context, arg sqlc.UpdateProfileParams) error
	DeleteProfile(ctx context.Context, id uuid.UUID) error
	UpdateProfileVerificationStatus(ctx context.Context, arg sqlc.UpdateProfileVerificationStatusParams) error
}

type ProfileService struct {
	repo      ProfileRepository
	log       *slog.Logger
	publisher *rabbit.Publisher
}

func NewProfileService(repo ProfileRepository, log *slog.Logger, publisher *rabbit.Publisher) *ProfileService {
	return &ProfileService{
		repo:      repo,
		log:       log,
		publisher: publisher,
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

func (s *ProfileService) UpdateVerificationStatus(ctx context.Context, userID uuid.UUID, verificationStatus string) error {
	verificationStatusValue := sqlc.NullVerificationStatus{}
	if err := verificationStatusValue.Scan(verificationStatus); err != nil {
		return err
	}
	arg := sqlc.UpdateProfileVerificationStatusParams{
		ID:                 userID,
		VerificationStatus: verificationStatusValue.VerificationStatus,
	}
	if verificationStatus == "verified" {
		go s.publishEvent(ProfileVerificationStatusUpdatedSubmit, userID)
	} else if verificationStatus == "rejected" {
		go s.publishEvent(ProfileVerificationStatusUpdatedReject, userID)
	}

	return s.repo.UpdateProfileVerificationStatus(ctx, arg)
}

func (s *ProfileService) publishEvent(event ProfileEventType, userID uuid.UUID) {
	profile, err := s.repo.GetProfile(context.Background(), userID)
	if err != nil {
		return
	}
	msg := ProfileEvent{
		Event:              event,
		UserID:             userID.String(),
		Email:              profile.Email,
		Name:               profile.Name,
		VerificationStatus: string(profile.VerificationStatus),
	}
	s.publisher.Publish("profile_queue", msg)
}
