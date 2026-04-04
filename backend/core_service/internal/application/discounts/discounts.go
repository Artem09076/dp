package discounts

import (
	"context"
	"log/slog"

	"github.com/Artem09076/dp/backend/core_service/internal/presentation/discounts/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/google/uuid"
)

type DiscountRepository interface {
	CreateDiscount(ctx context.Context, arg sqlc.CreateDiscountParams) (sqlc.Discount, error)
	GetService(ctx context.Context, id uuid.UUID) (sqlc.Service, error)
	GetDiscountById(ctx context.Context, id uuid.UUID) (sqlc.Discount, error)
	UpdateDiscount(ctx context.Context, arg sqlc.UpdateDiscountParams) error
	DeleteDiscount(ctx context.Context, id uuid.UUID) error
}

type DiscountService struct {
	repo DiscountRepository
	log  *slog.Logger
}

func NewDiscountService(repo DiscountRepository, log *slog.Logger) *DiscountService {
	return &DiscountService{
		repo: repo,
		log:  log,
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
	res, err := s.repo.CreateDiscount(ctx, params)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *DiscountService) GetDiscount(ctx context.Context, discountID uuid.UUID) (*sqlc.Discount, error) {
	res, err := s.repo.GetDiscountById(ctx, discountID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *DiscountService) UpdateDiscount(ctx context.Context, discountID uuid.UUID, updateDiscountObj dto.PatchDiscountRequest) error {
	discount, err := s.repo.GetDiscountById(ctx, discountID)
	if err != nil {
		return err
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
	return nil
}

func (s *DiscountService) DeleteDiscount(ctx context.Context, discountID uuid.UUID) error {
	if err := s.repo.DeleteDiscount(ctx, discountID); err != nil {
		return err
	}
	return nil
}
