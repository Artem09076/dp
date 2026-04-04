package reviews

import (
	"context"
	"database/sql"
	"log/slog"

	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/reviews/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
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
	repo ReviewsRepository
	log  *slog.Logger
}

func NewReviewService(repo ReviewsRepository, log *slog.Logger) *ReviewService {
	return &ReviewService{
		repo: repo,
		log:  log,
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, userID uuid.UUID, req dto.CreateReviewRequest) (*sqlc.Review, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, handlerlib.ErrInvalidInput
	}

	booking, err := s.repo.GetBookingByID(ctx, req.BookingID)
	if err != nil {
		return nil, handlerlib.ErrNotFound
	}

	if booking.ClientID != userID {
		return nil, handlerlib.ErrForbidden
	}

	if booking.Status != "completed" && booking.Status != "cancelled" {
		return nil, handlerlib.ErrInvalidInput
	}

	_, err = s.repo.GetReviewByBookingID(ctx, req.BookingID)
	if err == nil {
		return nil, handlerlib.ErrInvalidInput
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
	return &review, nil
}

func (s *ReviewService) GetReviewByBookingID(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.Review, error) {
	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, handlerlib.ErrNotFound
	}

	if booking.ClientID != userID {
		return nil, handlerlib.ErrForbidden
	}

	review, err := s.repo.GetReviewByBookingID(ctx, bookingID)
	if err != nil {
		return nil, handlerlib.ErrNotFound
	}
	return &review, nil
}

func (s *ReviewService) GetReviewsByServiceID(ctx context.Context, serviceID uuid.UUID, page int, limit int) ([]sqlc.Review, error) {
	reviews, err := s.repo.GetReviewsByServiceID(ctx, sqlc.GetReviewsByServiceIDParams{
		ServiceID: serviceID,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (s *ReviewService) PatchReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, req dto.PatchReviewRequest) error {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return handlerlib.ErrNotFound
	}
	arg := sqlc.UpdateReviewParams{
		ID:      reviewID,
		Rating:  review.Rating,
		Comment: review.Comment,
	}
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return handlerlib.ErrInvalidInput
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
		return handlerlib.ErrNotFound
	}

	if booking.ClientID != userID {
		return handlerlib.ErrForbidden
	}

	err = s.repo.UpdateReview(ctx, arg)
	if err != nil {
		return err
	}

	return nil

}

func (s *ReviewService) DeleteReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) error {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return handlerlib.ErrNotFound
	}

	booking, err := s.repo.GetBookingByID(ctx, review.BookingID)
	if err != nil {
		return handlerlib.ErrNotFound
	}

	if booking.ClientID != userID {
		return handlerlib.ErrForbidden
	}

	err = s.repo.DeleteReview(ctx, reviewID)
	if err != nil {
		return err
	}

	return nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID uuid.UUID) (*sqlc.Review, error) {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, handlerlib.ErrNotFound
	}
	return &review, nil
}
