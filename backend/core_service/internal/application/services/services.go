package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/services/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/redis"
	"github.com/google/uuid"
)

type ServiceRepository interface {
	CreateService(ctx context.Context, arg sqlc.CreateServiceParams) (sqlc.Service, error)
	SearchServices(ctx context.Context, arg sqlc.SearchServicesParams) ([]sqlc.Service, error)
	GetService(ctx context.Context, id uuid.UUID) (sqlc.GetServiceRow, error)
	GetServices(ctx context.Context, performerID uuid.UUID) ([]sqlc.Service, error)
	DeleteService(ctx context.Context, id uuid.UUID) error
	UpdateService(ctx context.Context, arg sqlc.UpdateServiceParams) error
	GetProfile(ctx context.Context, id uuid.UUID) (sqlc.GetProfileRow, error)
}

type Service struct {
	repo  ServiceRepository
	log   *slog.Logger
	redis *redis.RedisClient
}

func NewService(repo ServiceRepository, log *slog.Logger, redis *redis.RedisClient) *Service {
	return &Service{
		repo:  repo,
		log:   log,
		redis: redis,
	}
}

func (s *Service) CheckServiceOwnership(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID) (bool, error) {
	service, err := s.GetService(ctx, serviceID)
	if err != nil {
		s.log.Error("Failed to get service", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
		return false, err
	}
	user, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get service", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
		return false, err
	}
	return service.PerformerID == userID || user.Role == "admin", nil
}

func (s *Service) CreateService(ctx context.Context, createServiceObject dto.CreateServiceRequest) (*sqlc.Service, error) {
	performer, err := s.repo.GetProfile(ctx, createServiceObject.PerformerID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	if performer.VerificationStatus != "verified" {
		return nil, apierrors.ErrPerformerNotVerified
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

	go s.invalidateServiceCaches(context.Background(), res.ID.String(), createServiceObject.PerformerID.String())

	return &res, nil
}

func (s *Service) SearchServices(ctx context.Context, query string, page int, limit int) ([]sqlc.Service, error) {
	cachedResults, err := s.redis.SearchServices(ctx, query, page, limit)
	if err != nil {
		s.log.Warn("failed to get search results from cache", "error", err)
	}

	if cachedResults != nil {
		var services []sqlc.Service
		if err := json.Unmarshal(cachedResults, &services); err == nil {
			s.log.Debug("search results retrieved from cache", "query", query, "page", page)
			return services, nil
		}
	}

	validQuery := sql.NullString{
		String: query,
		Valid:  query != "",
	}
	param := sqlc.SearchServicesParams{
		Column1: validQuery,
		Limit:   int32(limit),
		Offset:  int32((page - 1) * limit),
	}

	services, err := s.repo.SearchServices(ctx, param)
	if err != nil {
		return nil, err
	}

	if err := s.redis.SetSearchServices(ctx, query, page, limit, services, 10*time.Minute); err != nil {
		s.log.Warn("failed to cache search results", "error", err)
	}

	return services, nil
}

func (s *Service) GetService(ctx context.Context, serviceID uuid.UUID) (*sqlc.GetServiceRow, error) {
	cachedService, err := s.redis.GetService(ctx, serviceID.String())
	if err != nil {
		s.log.Warn("failed to get service from cache", "error", err)
	}

	if cachedService != nil {
		serviceData, _ := json.Marshal(cachedService)
		var service sqlc.GetServiceRow
		if err := json.Unmarshal(serviceData, &service); err == nil {
			s.log.Debug("service retrieved from cache", "service_id", serviceID)
			return &service, nil
		}
	}

	res, err := s.repo.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	if err := s.redis.SetService(ctx, serviceID.String(), res, 1*time.Hour); err != nil {
		s.log.Warn("failed to cache service", "error", err)
	}

	return &res, nil
}

func (s *Service) GetServices(ctx context.Context, userID uuid.UUID) ([]sqlc.Service, error) {
	cachedServices, err := s.redis.GetServicesList(ctx, userID.String())
	if err != nil {
		s.log.Warn("failed to get services list from cache", "error", err)
	}

	if cachedServices != nil {
		var services []sqlc.Service
		if err := json.Unmarshal(cachedServices, &services); err == nil {
			s.log.Debug("services list retrieved from cache", "performer_id", userID)
			return services, nil
		}
	}

	services, err := s.repo.GetServices(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.redis.SetServicesList(ctx, userID.String(), services, 10*time.Minute); err != nil {
		s.log.Warn("failed to cache services list", "error", err)
	}

	return services, nil
}

func (s *Service) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	service, err := s.GetService(ctx, serviceID)
	if err != nil {
		return err
	}

	err = s.repo.DeleteService(ctx, serviceID)
	if err != nil {
		return err
	}

	go s.invalidateServiceCaches(context.Background(), serviceID.String(), service.PerformerID.String())

	return nil
}

func (s *Service) UpdateService(ctx context.Context, serviceID uuid.UUID, updateServiceObject dto.PatchServiceRequest) error {
	service, err := s.GetService(ctx, serviceID)
	if err != nil {
		return apierrors.ErrNotFound
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

	go s.invalidateServiceCaches(context.Background(), serviceID.String(), service.PerformerID.String())

	return nil
}

func (s *Service) invalidateServiceCaches(ctx context.Context, serviceID, performerID string) {
	s.redis.InvalidateService(ctx, serviceID)
	s.redis.InvalidateServicesList(ctx, performerID)
	s.redis.InvalidateSearchServices(ctx)
	s.redis.InvalidateServiceDiscounts(ctx, serviceID)
	s.redis.InvalidateServiceReviews(ctx, serviceID)
}
