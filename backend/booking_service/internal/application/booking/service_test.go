package booking

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"

	"github.com/Artem09076/dp/backend/booking_service/internal/lib/api/errors"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) WithTx(tx *sql.Tx) *sqlc.Queries {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sqlc.Queries)
}

func (m *MockRepository) CreateBooking(ctx context.Context, booking sqlc.CreateBookingParams) (uuid.UUID, error) {
	args := m.Called(ctx, booking)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRepository) CancelBooking(ctx context.Context, bookingID uuid.UUID) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockRepository) GetServiceById(ctx context.Context, id uuid.UUID) (sqlc.Service, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Service), args.Error(1)
}

func (m *MockRepository) GetDiscountById(ctx context.Context, id uuid.UUID) (sqlc.Discount, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Discount), args.Error(1)
}

func (m *MockRepository) GetBookingByID(ctx context.Context, id uuid.UUID) (sqlc.GetBookingByIDRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.GetBookingByIDRow), args.Error(1)
}

func (m *MockRepository) GetUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockRepository) SubmitBooking(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) UpdateBookingTime1(ctx context.Context, arg sqlc.UpdateBookingTime1Params) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRepository) UpdateBookingTime2(ctx context.Context, arg sqlc.UpdateBookingTime2Params) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRepository) GetBookingByClientID(ctx context.Context, clientID uuid.UUID) ([]sqlc.Booking, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.Booking), args.Error(1)
}

func (m *MockRepository) GetBookingByPerformerID(ctx context.Context, performerID uuid.UUID) ([]sqlc.Booking, error) {
	args := m.Called(ctx, performerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.Booking), args.Error(1)
}

func (m *MockRepository) ServiceExists(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) IncreaseDiscountUsage(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteBooking(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetBookingsForUpdate(ctx context.Context, arg sqlc.GetBookingsForUpdateParams) ([]sqlc.GetBookingsForUpdateRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.GetBookingsForUpdateRow), args.Error(1)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(queue string, msg interface{}) error {
	args := m.Called(queue, msg)
	return args.Error(0)
}

var (
	testServiceID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testDiscountID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testPerformerID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testClientID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testAdminID     = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	testBookingID   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

func setupTestService(t *testing.T) (*MockRepository, *MockPublisher, *BookingService) {
	mockRepo := new(MockRepository)
	mockPublisher := new(MockPublisher)
	logger := slog.Default()
	service := &BookingService{
		repo:      mockRepo,
		db:        nil,
		log:       logger,
		publisher: mockPublisher,
	}
	return mockRepo, mockPublisher, service
}

func TestCalculateFinalPrice_Percentage(t *testing.T) {
	service := &BookingService{}

	testCases := []struct {
		name      string
		basePrice int32
		discount  sqlc.Discount
		expected  int32
	}{
		{
			name:      "20% discount on 1000",
			basePrice: 1000,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypePercentage,
				Value: 20,
			},
			expected: 800,
		},
		{
			name:      "50% discount on 500",
			basePrice: 500,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypePercentage,
				Value: 50,
			},
			expected: 250,
		},
		{
			name:      "0% discount",
			basePrice: 1000,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypePercentage,
				Value: 0,
			},
			expected: 1000,
		},
		{
			name:      "100% discount",
			basePrice: 1000,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypePercentage,
				Value: 100,
			},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.calculateFinalPrice(tc.basePrice, tc.discount)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateFinalPrice_FixedAmount(t *testing.T) {
	service := &BookingService{}

	testCases := []struct {
		name      string
		basePrice int32
		discount  sqlc.Discount
		expected  int32
	}{
		{
			name:      "200 discount on 1000",
			basePrice: 1000,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypeFixedAmount,
				Value: 200,
			},
			expected: 800,
		},
		{
			name:      "500 discount on 500",
			basePrice: 500,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypeFixedAmount,
				Value: 500,
			},
			expected: 0,
		},
		{
			name:      "0 discount",
			basePrice: 1000,
			discount: sqlc.Discount{
				Type:  sqlc.DiscoutnTypeFixedAmount,
				Value: 0,
			},
			expected: 1000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.calculateFinalPrice(tc.basePrice, tc.discount)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckBookingOwnerships_AsClient(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	mockRepo.On("GetUserById", ctx, testClientID).Return(sqlc.User{
		ID:   testClientID,
		Role: "client",
	}, nil)

	result := service.CheckBookingOwnerships(ctx, testClientID, testClientID, testPerformerID)

	assert.True(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCheckBookingOwnerships_AsPerformer(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	mockRepo.On("GetUserById", ctx, testPerformerID).Return(sqlc.User{
		ID:   testPerformerID,
		Role: "performer",
	}, nil)

	result := service.CheckBookingOwnerships(ctx, testPerformerID, testClientID, testPerformerID)

	assert.True(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCheckBookingOwnerships_AsAdmin(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	mockRepo.On("GetUserById", ctx, testAdminID).Return(sqlc.User{
		ID:   testAdminID,
		Role: "admin",
	}, nil)

	result := service.CheckBookingOwnerships(ctx, testAdminID, testClientID, testPerformerID)

	assert.True(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCheckBookingOwnerships_NoAccess(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	otherUserID := uuid.New()

	mockRepo.On("GetUserById", ctx, otherUserID).Return(sqlc.User{
		ID:   otherUserID,
		Role: "client",
	}, nil)

	result := service.CheckBookingOwnerships(ctx, otherUserID, testClientID, testPerformerID)

	assert.False(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCheckBookingOwnerships_UserNotFound(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	otherUserID := uuid.New()

	mockRepo.On("GetUserById", ctx, otherUserID).Return(sqlc.User{}, assert.AnError)

	result := service.CheckBookingOwnerships(ctx, otherUserID, testClientID, testPerformerID)

	assert.False(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCancelBooking_Forbidden(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	otherUserID := uuid.New()
	booking := &sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
	}

	mockRepo.On("GetUserById", ctx, otherUserID).Return(sqlc.User{
		ID:   otherUserID,
		Role: "client",
	}, nil)

	err := service.CancelBooking(ctx, otherUserID, booking)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrForbidden, err)
	mockRepo.AssertExpectations(t)
}

func TestSubmitBooking_AlreadyConfirmed(t *testing.T) {
	_, _, service := setupTestService(t)
	ctx := context.Background()

	booking := &sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
		Status:      "confirmed",
	}

	err := service.SubmitBooking(ctx, testPerformerID, booking)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrAlreadyDone, err)
}

func TestSubmitBooking_NotPerformer(t *testing.T) {
	_, _, service := setupTestService(t)
	ctx := context.Background()

	booking := &sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
		Status:      "pending",
	}

	err := service.SubmitBooking(ctx, testClientID, booking)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrForbidden, err)
}

func TestGetBooking_Success(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	expectedBooking := sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
	}

	mockRepo.On("GetBookingByID", ctx, testBookingID).Return(expectedBooking, nil)
	mockRepo.On("GetUserById", ctx, testClientID).Return(sqlc.User{
		ID:   testClientID,
		Role: "client",
	}, nil)

	booking, err := service.GetBooking(ctx, testClientID, testBookingID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBooking.ID, booking.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetBooking_NotFound(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	mockRepo.On("GetBookingByID", ctx, testBookingID).Return(sqlc.GetBookingByIDRow{}, assert.AnError)

	booking, err := service.GetBooking(ctx, testClientID, testBookingID)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
	assert.Nil(t, booking)
}

func TestGetBooking_Forbidden(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	otherUserID := uuid.New()
	bookingRow := sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
	}

	mockRepo.On("GetBookingByID", ctx, testBookingID).Return(bookingRow, nil)
	mockRepo.On("GetUserById", ctx, otherUserID).Return(sqlc.User{
		ID:   otherUserID,
		Role: "client",
	}, nil)

	booking, err := service.GetBooking(ctx, otherUserID, testBookingID)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrForbidden, err)
	assert.Nil(t, booking)
}

func TestGetBookings_AsClient(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	expectedBookings := []sqlc.Booking{
		{ID: testBookingID, ClientID: testClientID},
	}

	mockRepo.On("GetUserById", ctx, testClientID).Return(sqlc.User{
		ID:   testClientID,
		Role: "client",
	}, nil)
	mockRepo.On("GetBookingByClientID", ctx, testClientID).Return(expectedBookings, nil)

	bookings, err := service.GetBookings(ctx, testClientID)

	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
	assert.Equal(t, testBookingID, bookings[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestGetBookings_AsPerformer(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	expectedBookings := []sqlc.Booking{
		{ID: testBookingID, ClientID: testClientID},
	}

	mockRepo.On("GetUserById", ctx, testPerformerID).Return(sqlc.User{
		ID:   testPerformerID,
		Role: "performer",
	}, nil)
	mockRepo.On("GetBookingByPerformerID", ctx, testPerformerID).Return(expectedBookings, nil)

	bookings, err := service.GetBookings(ctx, testPerformerID)

	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
	mockRepo.AssertExpectations(t)
}

func TestGetBookings_InvalidRole(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	invalidUserID := uuid.New()

	mockRepo.On("GetUserById", ctx, invalidUserID).Return(sqlc.User{
		ID:   invalidUserID,
		Role: "invalid",
	}, nil)

	bookings, err := service.GetBookings(ctx, invalidUserID)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrForbidden, err)
	assert.Nil(t, bookings)
}

func TestDeleteBooking_Success(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	booking := sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
	}

	mockRepo.On("GetBookingByID", ctx, testBookingID).Return(booking, nil)
	mockRepo.On("GetUserById", ctx, testClientID).Return(sqlc.User{
		ID:   testClientID,
		Role: "client",
	}, nil)
	mockRepo.On("DeleteBooking", ctx, testBookingID).Return(nil)

	err := service.DeleteBooking(ctx, testClientID, testBookingID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteBooking_NotFound(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	mockRepo.On("GetBookingByID", ctx, testBookingID).Return(sqlc.GetBookingByIDRow{}, assert.AnError)

	err := service.DeleteBooking(ctx, testClientID, testBookingID)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestDeleteBooking_Forbidden(t *testing.T) {
	mockRepo, _, service := setupTestService(t)
	ctx := context.Background()

	otherUserID := uuid.New()
	booking := sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
	}

	mockRepo.On("GetBookingByID", ctx, testBookingID).Return(booking, nil)
	mockRepo.On("GetUserById", ctx, otherUserID).Return(sqlc.User{
		ID:   otherUserID,
		Role: "client",
	}, nil)

	err := service.DeleteBooking(ctx, otherUserID, testBookingID)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrForbidden, err)
}
