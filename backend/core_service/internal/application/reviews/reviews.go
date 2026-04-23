package reviews

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/reviews/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/redis"
	"github.com/google/uuid"
)

type ReviewsRepository interface {
	CreateReview(ctx context.Context, arg sqlc.CreateReviewParams) (sqlc.Review, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (sqlc.GetBookingByIDRow, error)
	GetReviewByID(ctx context.Context, id uuid.UUID) (sqlc.Review, error)
	GetReviewByBookingID(ctx context.Context, bookingID uuid.UUID) (sqlc.Review, error)
	GetReviewsByServiceID(ctx context.Context, arg sqlc.GetReviewsByServiceIDParams) ([]sqlc.Review, error)
	UpdateReview(ctx context.Context, arg sqlc.UpdateReviewParams) error
	DeleteReview(ctx context.Context, id uuid.UUID) error
}

type ReviewService struct {
	repo  ReviewsRepository
	log   *slog.Logger
	redis *redis.RedisClient
}

func NewReviewService(repo ReviewsRepository, log *slog.Logger, redis *redis.RedisClient) *ReviewService {
	return &ReviewService{
		repo:  repo,
		log:   log,
		redis: redis,
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, userID uuid.UUID, req dto.CreateReviewRequest) (*sqlc.Review, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, apierrors.ErrInvalidInput
	}

	booking, err := s.repo.GetBookingByID(ctx, req.BookingID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}

	if booking.ClientID != userID {
		return nil, apierrors.ErrForbidden
	}

	if booking.Status != "completed" && booking.Status != "cancelled" {
		return nil, apierrors.ErrInvalidReviewStatus
	}

	_, err = s.repo.GetReviewByBookingID(ctx, req.BookingID)
	if err == nil {
		return nil, apierrors.ErrReviewAlreadyExists
	}

	var comment sql.NullString
	if req.Comment != "" {
		comment = sql.NullString{
			String: req.Comment,
			Valid:  true,
		}
	} else {
		comment = sql.NullString{
			Valid: false,
		}
	}

	review, err := s.repo.CreateReview(ctx, sqlc.CreateReviewParams{
		BookingID: req.BookingID,
		Rating:    req.Rating,
		Comment:   comment,
	})
	if err != nil {
		return nil, err
	}

	go s.invalidateReviewCaches(context.Background(), review.ID.String(), booking.ServiceID.String())

	return &review, nil
}

func (s *ReviewService) GetReviewByBookingID(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.Review, error) {
	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		s.log.Info(err.Error())
		return nil, apierrors.ErrNotFound
	}

	if booking.ClientID != userID {
		return nil, apierrors.ErrForbidden
	}

	review, err := s.repo.GetReviewByBookingID(ctx, bookingID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return &review, nil
}

func (s *ReviewService) GetReviewsByServiceID(ctx context.Context, serviceID uuid.UUID, page int, limit int) ([]sqlc.Review, error) {
	cachedReviews, err := s.redis.GetServiceReviews(ctx, serviceID.String(), page, limit)
	if err != nil {
		s.log.Warn("failed to get service reviews from cache", "error", err)
	}

	if cachedReviews != nil {
		var reviews []sqlc.Review
		if err := json.Unmarshal(cachedReviews, &reviews); err == nil {
			s.log.Debug("service reviews retrieved from cache", "service_id", serviceID, "page", page)
			return reviews, nil
		}
	}

	reviews, err := s.repo.GetReviewsByServiceID(ctx, sqlc.GetReviewsByServiceIDParams{
		ServiceID: serviceID,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}

	if err := s.redis.SetServiceReviews(ctx, serviceID.String(), page, limit, reviews, 10*time.Minute); err != nil {
		s.log.Warn("failed to cache service reviews", "error", err)
	}

	return reviews, nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID uuid.UUID) (*sqlc.Review, error) {
	cachedReview, err := s.redis.GetReview(ctx, reviewID.String())
	if err != nil {
		s.log.Warn("failed to get review from cache", "error", err)
	}

	if cachedReview != nil {
		var review sqlc.Review
		if err := json.Unmarshal(cachedReview, &review); err == nil {
			s.log.Debug("review retrieved from cache", "review_id", reviewID)
			return &review, nil
		}
	}

	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}

	if err := s.redis.SetReview(ctx, reviewID.String(), review, 15*time.Minute); err != nil {
		s.log.Warn("failed to cache review", "error", err)
	}

	return &review, nil
}

func (s *ReviewService) PatchReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, req dto.PatchReviewRequest) error {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return apierrors.ErrNotFound
	}

	arg := sqlc.UpdateReviewParams{
		ID:      reviewID,
		Rating:  review.Rating,
		Comment: review.Comment,
	}
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return apierrors.ErrInvalidInput
		}
		arg.Rating = *req.Rating
	}
	if req.Comment != nil {
		if *req.Comment != "" {
			arg.Comment = sql.NullString{
				String: *req.Comment,
				Valid:  true,
			}
		} else {
			arg.Comment = sql.NullString{
				Valid: false,
			}
		}
	}

	booking, err := s.repo.GetBookingByID(ctx, review.BookingID)
	if err != nil {
		return apierrors.ErrNotFound
	}

	if booking.ClientID != userID {
		return apierrors.ErrForbidden
	}

	err = s.repo.UpdateReview(ctx, arg)
	if err != nil {
		return err
	}

	go s.invalidateReviewCaches(context.Background(), reviewID.String(), booking.ServiceID.String())

	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) error {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return apierrors.ErrNotFound
	}

	booking, err := s.repo.GetBookingByID(ctx, review.BookingID)
	if err != nil {
		return apierrors.ErrNotFound
	}

	if booking.ClientID != userID {
		return apierrors.ErrForbidden
	}

	err = s.repo.DeleteReview(ctx, reviewID)
	if err != nil {
		return err
	}

	go s.invalidateReviewCaches(context.Background(), reviewID.String(), booking.ServiceID.String())

	return nil
}

func (s *ReviewService) invalidateReviewCaches(ctx context.Context, reviewID, serviceID string) {
	s.redis.InvalidateReview(ctx, reviewID)
	s.redis.InvalidateServiceReviews(ctx, serviceID)
}
