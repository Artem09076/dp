package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Artem09076/dp/backend/core_service/internal/presentation/services/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
)

type ServiceRepository interface {
	CreateService(ctx context.Context, arg sqlc.CreateServiceParams) (sqlc.Service, error)
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

func (s *Service) CreateService(ctx context.Context, createServiceObject dto.CreateServiceRequest) (*sqlc.Service, error) {
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
