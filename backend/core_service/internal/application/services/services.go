package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Artem09076/dp/backend/core_service/internal/presentation/services/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/google/uuid"
)

type ServiceRepository interface {
	CreateService(ctx context.Context, arg sqlc.CreateServiceParams) (sqlc.Service, error)
	SearchServices(ctx context.Context, arg sqlc.SearchServicesParams) ([]sqlc.Service, error)
	GetService(ctx context.Context, id uuid.UUID) (sqlc.Service, error)
	DeleteService(ctx context.Context, id uuid.UUID) error
	UpdateService(ctx context.Context, arg sqlc.UpdateServiceParams) error
	GetProfile(ctx context.Context, id uuid.UUID) (sqlc.GetProfileRow, error)
}

type Service struct {
	repo ServiceRepository
	log  *slog.Logger
}

func NewService(repo ServiceRepository, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) CheckServiceOwnership(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID) (bool, error) {
	service, err := s.repo.GetService(ctx, serviceID)
	if err != nil {
		s.log.Error("Failed to get service", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
		return false, err
	}
	return service.PerformerID == userID, nil
}

func (s *Service) CreateService(ctx context.Context, createServiceObject dto.CreateServiceRequest) (*sqlc.Service, error) {
	performer, err := s.repo.GetProfile(ctx, createServiceObject.PerformerID)
	if err != nil {
		return nil, err
	}
	if performer.VerificationStatus != "verified" {
		return nil, err
		// Нормально распиши
	}
	var NullDescription sql.NullString
	if createServiceObject.Description == "" {
		NullDescription = sql.NullString{String: createServiceObject.Description, Valid: false}
	} else {
		NullDescription = sql.NullString{String: createServiceObject.Description, Valid: true}
	}
	param := sqlc.CreateServiceParams{
		PerformerID:     createServiceObject.PerformerID,
		Title:           createServiceObject.Title,
		Description:     NullDescription,
		Price:           int64(createServiceObject.Price),
		DurationMinutes: int32(createServiceObject.DurationMinutes),
	}
	res, err := s.repo.CreateService(ctx, param)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *Service) SearchServices(ctx context.Context, query string, page int, limit int) ([]sqlc.Service, error) {
	validQuery := sql.NullString{
		String: query,
		Valid:  query != "",
	}
	param := sqlc.SearchServicesParams{
		Column1: validQuery,
		Limit:   int32(limit),
		Offset:  int32((page - 1) * limit),
	}
	res, err := s.repo.SearchServices(ctx, param)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Service) GetService(ctx context.Context, serviceID uuid.UUID) (*sqlc.Service, error) {
	res, err := s.repo.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *Service) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	return s.repo.DeleteService(ctx, serviceID)
}

func (s *Service) UpdateService(ctx context.Context, serviceID uuid.UUID, updateServiceObject dto.PatchServiceRequest) error {
	service, err := s.repo.GetService(ctx, serviceID)
	if err != nil {
		return err
	}
	arg := sqlc.UpdateServiceParams{
		ID:              serviceID,
		Title:           service.Title,
		Description:     service.Description,
		Price:           service.Price,
		DurationMinutes: service.DurationMinutes,
	}
	if updateServiceObject.Title != nil {
		arg.Title = *updateServiceObject.Title
	}
	if updateServiceObject.Description != nil {
		arg.Description = sql.NullString{String: *updateServiceObject.Description, Valid: true}
	}
	if updateServiceObject.Price != nil {
		arg.Price = int64(*updateServiceObject.Price)
	}
	if updateServiceObject.DurationMinutes != nil {
		arg.DurationMinutes = int32(*updateServiceObject.DurationMinutes)
	}
	err = s.repo.UpdateService(ctx, arg)
	if err != nil {
		return err
	}
	return nil

}
