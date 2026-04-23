package admin

import (
	"context"
	"log/slog"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/admin/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/redis"
	"github.com/google/uuid"
)

type AdminRepository interface {
	GetUnverifiedPerformers(ctx context.Context, arg sqlc.GetUnverifiedPerformersParams) ([]sqlc.GetUnverifiedPerformersRow, error)
	CountUnverifiedPerformers(ctx context.Context) (int64, error)
	GetUsersWithFilters(ctx context.Context, arg sqlc.GetUsersWithFiltersParams) ([]sqlc.GetUsersWithFiltersRow, error)
	CountUsersWithFilters(ctx context.Context) (int64, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.GetUserByIDRow, error)
	UpdateUserRole(ctx context.Context, arg sqlc.UpdateUserRoleParams) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdateVerificationStatus(ctx context.Context, arg sqlc.UpdateVerificationStatusParams) error
	GetAllServices(ctx context.Context, arg sqlc.GetAllServicesParams) ([]sqlc.GetAllServicesRow, error)
	CountAllServices(ctx context.Context, performerID uuid.UUID) (int64, error)
	GetAllBookings(ctx context.Context, arg sqlc.GetAllBookingsParams) ([]sqlc.GetAllBookingsRow, error)
	CountAllBookings(ctx context.Context) (int64, error)
	GetAllReviews(ctx context.Context, arg sqlc.GetAllReviewsParams) ([]sqlc.GetAllReviewsRow, error)
	CountAllReviews(ctx context.Context) (int64, error)
	DeleteReview(ctx context.Context, id uuid.UUID) error
	GetService(ctx context.Context, id uuid.UUID) (sqlc.GetServiceRow, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (sqlc.GetBookingByIDRow, error)
	GetReviewByID(ctx context.Context, id uuid.UUID) (sqlc.Review, error)
}

type AdminService struct {
	repo  AdminRepository
	log   *slog.Logger
	redis *redis.RedisClient
}

func NewAdminService(repo AdminRepository, log *slog.Logger, redis *redis.RedisClient) *AdminService {
	return &AdminService{
		repo:  repo,
		log:   log,
		redis: redis,
	}
}

func (s *AdminService) GetUnverifiedPerformers(ctx context.Context, page, pageSize int) ([]dto.UserResponse, int64, error) {
	offset := (page - 1) * pageSize

	performers, err := s.repo.GetUnverifiedPerformers(ctx, sqlc.GetUnverifiedPerformersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountUnverifiedPerformers(ctx)
	if err != nil {
		return nil, 0, err
	}

	users := make([]dto.UserResponse, len(performers))
	for i, p := range performers {
		user := dto.UserResponse{
			ID:                 p.ID.String(),
			Name:               p.Name,
			Email:              p.Email,
			Role:               string(p.Role),
			VerificationStatus: string(p.VerificationStatus),
			CreatedAt:          p.CreatedAt,
			UpdatedAt:          p.UpdatedAt,
			ServicesCount:      p.ServicesCount,
			TotalBookings:      p.TotalBookings,
		}
		if p.Inn.Valid {
			user.INN = &p.Inn.String
		}
		if p.BusinessType.Valid {
			businessType := string(p.BusinessType.BusinessType)
			user.BusinessType = &businessType
		}
		users[i] = user
	}

	return users, total, nil
}

func (s *AdminService) GetUsers(ctx context.Context, role, verificationStatus, search string, page, pageSize int) ([]dto.UserResponse, int64, error) {
	offset := (page - 1) * pageSize

	var roleEnum sqlc.NullUserRole
	if role != "" {
		roleEnum.Scan(role)
	}

	var verificationEnum sqlc.NullVerificationStatus
	if verificationStatus != "" {
		verificationEnum.Scan(verificationStatus)
	}

	users, err := s.repo.GetUsersWithFilters(ctx, sqlc.GetUsersWithFiltersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountUsersWithFilters(ctx)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.UserResponse, len(users))
	for i, u := range users {
		user := dto.UserResponse{
			ID:                 u.ID.String(),
			Name:               u.Name,
			Email:              u.Email,
			Role:               string(u.Role),
			VerificationStatus: string(u.VerificationStatus),
			CreatedAt:          u.CreatedAt,
			UpdatedAt:          u.UpdatedAt,
		}
		if u.Inn.Valid {
			user.INN = &u.Inn.String
		}
		if u.BusinessType.Valid {
			businessType := string(u.BusinessType.BusinessType)
			user.BusinessType = &businessType
		}
		responses[i] = user
	}

	return responses, total, nil
}

func (s *AdminService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}

	resp := &dto.UserResponse{
		ID:                 user.ID.String(),
		Name:               user.Name,
		Email:              user.Email,
		Role:               string(user.Role),
		VerificationStatus: string(user.VerificationStatus),
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
	if user.Inn.Valid {
		resp.INN = &user.Inn.String
	}
	if user.BusinessType.Valid {
		businessType := string(user.BusinessType.BusinessType)
		resp.BusinessType = &businessType
	}

	return resp, nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	var roleEnum sqlc.NullUserRole
	if err := roleEnum.Scan(role); err != nil {
		return apierrors.ErrInvalidInput
	}

	err := s.repo.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		ID:   userID,
		Role: roleEnum.UserRole,
	})
	if err != nil {
		return err
	}

	go s.redis.InvalidateProfile(context.Background(), userID.String())
	return nil
}

func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	go s.redis.InvalidateProfile(context.Background(), userID.String())
	return s.repo.DeleteUser(ctx, userID)
}

func (s *AdminService) BatchVerifyPerformers(ctx context.Context, userIDs []string, status string) error {
	for _, idStr := range userIDs {
		userID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}

		var verificationStatus sqlc.NullVerificationStatus
		if err := verificationStatus.Scan(status); err != nil {
			continue
		}

		err = s.repo.UpdateVerificationStatus(ctx, sqlc.UpdateVerificationStatusParams{
			ID:                 userID,
			VerificationStatus: verificationStatus.VerificationStatus,
		})
		if err != nil {
			s.log.Error("failed to update verification status", "user_id", userID, "error", err)
			continue
		}

		go s.redis.InvalidateProfile(context.Background(), userID.String())
	}

	return nil
}

func (s *AdminService) GetServices(ctx context.Context, performerID uuid.UUID, page, pageSize int) ([]dto.ServiceResponse, int64, error) {
	offset := (page - 1) * pageSize

	services, err := s.repo.GetAllServices(ctx, sqlc.GetAllServicesParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountAllServices(ctx, performerID)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.ServiceResponse, len(services))
	for i, srv := range services {
		resp := dto.ServiceResponse{
			ID:              srv.ID.String(),
			PerformerID:     srv.PerformerID.String(),
			PerformerName:   srv.PerformerName,
			Title:           srv.Title,
			Price:           srv.Price,
			DurationMinutes: srv.DurationMinutes,
			CreatedAt:       srv.CreatedAt,
			UpdatedAt:       srv.UpdatedAt,
		}
		if srv.Description.Valid {
			resp.Description = &srv.Description.String
		}
		responses[i] = resp
	}

	return responses, total, nil
}

func (s *AdminService) GetBookings(ctx context.Context, status, clientID, performerID string, page, pageSize int) ([]dto.BookingResponse, int64, error) {
	offset := (page - 1) * pageSize

	var statusEnum sqlc.NullBookingStatus
	if status != "" {
		statusEnum.Scan(status)
	}

	bookings, err := s.repo.GetAllBookings(ctx, sqlc.GetAllBookingsParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountAllBookings(ctx)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.BookingResponse, len(bookings))
	for i, b := range bookings {
		responses[i] = dto.BookingResponse{
			ID:          b.ID.String(),
			ClientID:    b.ClientID.String(),
			ClientName:  b.ClientName,
			ServiceID:   b.ServiceID.String(),
			ServiceName: b.ServiceName,
			BasePrice:   int64(b.BasePrice),
			FinalPrice:  int64(b.FinalPrice),
			BookingTime: b.BookingTime,
			Status:      string(b.Status),
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		}
	}

	return responses, total, nil
}

func (s *AdminService) GetReviews(ctx context.Context, serviceID uuid.UUID, rating int, page, pageSize int) ([]dto.ReviewResponse, int64, error) {
	offset := (page - 1) * pageSize

	reviews, err := s.repo.GetAllReviews(ctx, sqlc.GetAllReviewsParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountAllReviews(ctx)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.ReviewResponse, len(reviews))
	for i, r := range reviews {
		resp := dto.ReviewResponse{
			ID:         r.ID.String(),
			BookingID:  r.BookingID.String(),
			Rating:     r.Rating,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
			ClientName: r.ClientName,
			ServiceID:  r.ServiceID.String(),
		}
		if r.Comment.Valid {
			resp.Comment = &r.Comment.String
		}
		responses[i] = resp
	}

	return responses, total, nil
}

func (s *AdminService) DeleteReview(ctx context.Context, reviewID uuid.UUID) error {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return apierrors.ErrNotFound
	}

	booking, err := s.repo.GetBookingByID(ctx, review.BookingID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteReview(ctx, reviewID); err != nil {
		return err
	}

	go s.redis.InvalidateReview(context.Background(), reviewID.String())
	go s.redis.InvalidateServiceReviews(context.Background(), booking.ServiceID.String())

	return nil
}
