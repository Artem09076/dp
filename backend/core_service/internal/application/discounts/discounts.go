package discounts

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/discounts/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/redis"
	"github.com/google/uuid"
)

type DiscountRepository interface {
	CreateDiscount(ctx context.Context, arg sqlc.CreateDiscountParams) (sqlc.Discount, error)
	GetService(ctx context.Context, id uuid.UUID) (sqlc.GetServiceRow, error)
	GetDiscountById(ctx context.Context, id uuid.UUID) (sqlc.Discount, error)
	UpdateDiscount(ctx context.Context, arg sqlc.UpdateDiscountParams) error
	DeleteDiscount(ctx context.Context, id uuid.UUID) error
	GetDiscountsByServiceID(ctx context.Context, serviceID uuid.UUID) ([]sqlc.Discount, error)
}

type DiscountService struct {
	repo  DiscountRepository
	log   *slog.Logger
	redis *redis.RedisClient
}

func NewDiscountService(repo DiscountRepository, log *slog.Logger, redis *redis.RedisClient) *DiscountService {
	return &DiscountService{
		repo:  repo,
		log:   log,
		redis: redis,
	}
}

func (s *DiscountService) CheckServiceOwnership(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID) (bool, error) {
	service, err := s.repo.GetService(ctx, serviceID)
	if err != nil {
		s.log.Error("Failed to get service", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
		return false, err
	}
	return service.PerformerID == userID, nil
}

func (s *DiscountService) CreateDiscount(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID, params sqlc.CreateDiscountParams) (*sqlc.Discount, error) {
	owns, err := s.CheckServiceOwnership(ctx, userID, serviceID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, apierrors.ErrForbidden
	}

	res, err := s.repo.CreateDiscount(ctx, params)
	if err != nil {
		return nil, err
	}

	go s.invalidateDiscountCaches(context.Background(), res.ID.String(), serviceID.String())

	return &res, nil
}

func (s *DiscountService) GetDiscount(ctx context.Context, discountID uuid.UUID) (*sqlc.Discount, error) {
	cachedDiscount, err := s.redis.GetDiscount(ctx, discountID.String())
	if err != nil {
		s.log.Warn("failed to get discount from cache", "error", err)
	}

	if cachedDiscount != nil {
		var discount sqlc.Discount
		if err := json.Unmarshal(cachedDiscount, &discount); err == nil {
			s.log.Debug("discount retrieved from cache", "discount_id", discountID)
			return &discount, nil
		}
	}

	res, err := s.repo.GetDiscountById(ctx, discountID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}

	if err := s.redis.SetDiscount(ctx, discountID.String(), res, 30*time.Minute); err != nil {
		s.log.Warn("failed to cache discount", "error", err)
	}

	return &res, nil
}

func (s *DiscountService) GetDiscountsByServiceID(ctx context.Context, serviceID uuid.UUID) ([]sqlc.Discount, error) {

	cachedDiscounts, err := s.redis.GetServiceDiscounts(ctx, serviceID.String())
	if err != nil {
		s.log.Warn("failed to get service discounts from cache", "error", err, "service_id", serviceID)
	}

	if cachedDiscounts != nil {
		var discounts []sqlc.Discount
		if err := json.Unmarshal(cachedDiscounts, &discounts); err == nil {
			s.log.Debug("discounts retrieved from cache", "service_id", serviceID, "count", len(discounts))
			return discounts, nil
		}
	}

	discounts, err := s.repo.GetDiscountsByServiceID(ctx, serviceID)
	if err != nil {
		s.log.Error("failed to get discounts by service ID", "error", err, "service_id", serviceID)
		return nil, err
	}

	if err := s.redis.SetServiceDiscounts(ctx, serviceID.String(), discounts, 10*time.Minute); err != nil {
		s.log.Warn("failed to cache service discounts", "error", err, "service_id", serviceID)
	}

	s.log.Debug("discounts retrieved from database", "service_id", serviceID, "count", len(discounts))
	return discounts, nil
}
func (s *DiscountService) UpdateDiscount(ctx context.Context, discountID uuid.UUID, updateDiscountObj dto.PatchDiscountRequest) error {
	discount, err := s.repo.GetDiscountById(ctx, discountID)
	if err != nil {
		return err
	}

	if updateDiscountObj.ValidFrom != nil && updateDiscountObj.ValidTo != nil {
		if updateDiscountObj.ValidFrom.After(*updateDiscountObj.ValidTo) {
			return apierrors.ErrInvalidInput
		}
	}

	param := sqlc.UpdateDiscountParams{
		ID:        discountID,
		ValidFrom: discount.ValidFrom,
		ValidTo:   discount.ValidTo,
		MaxUses:   discount.MaxUses,
	}
	if updateDiscountObj.ValidFrom != nil {
		param.ValidFrom = *updateDiscountObj.ValidFrom
	}
	if updateDiscountObj.ValidTo != nil {
		param.ValidTo = *updateDiscountObj.ValidTo
	}
	if updateDiscountObj.MaxUses != nil {
		param.MaxUses = int32(*updateDiscountObj.MaxUses)
	}

	if err := s.repo.UpdateDiscount(ctx, param); err != nil {
		return err
	}

	go s.invalidateDiscountCaches(context.Background(), discountID.String(), discount.ServiceID.String())

	return nil
}

func (s *DiscountService) DeleteDiscount(ctx context.Context, discountID uuid.UUID) error {
	discount, err := s.repo.GetDiscountById(ctx, discountID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteDiscount(ctx, discountID); err != nil {
		return err
	}

	go s.invalidateDiscountCaches(context.Background(), discountID.String(), discount.ServiceID.String())

	return nil
}

func (s *DiscountService) invalidateDiscountCaches(ctx context.Context, discountID, serviceID string) {
	s.redis.InvalidateDiscount(ctx, discountID)
	s.redis.InvalidateServiceDiscounts(ctx, serviceID)
}
