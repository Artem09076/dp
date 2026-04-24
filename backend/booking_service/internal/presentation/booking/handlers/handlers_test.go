package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apierrors "github.com/Artem09076/dp/backend/booking_service/internal/lib/api/errors"

	"github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/dto"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBookingService struct {
	mock.Mock
}

func (m *MockBookingService) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.CreateBookingRequest) (uuid.UUID, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockBookingService) CancelBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	args := m.Called(ctx, userID, booking)
	return args.Error(0)
}

func (m *MockBookingService) SubmitBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	args := m.Called(ctx, userID, booking)
	return args.Error(0)
}

func (m *MockBookingService) CompleteBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	args := m.Called(ctx, userID, booking)
	return args.Error(0)
}

func (m *MockBookingService) UpdateBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req dto.PatchBookingRequest) error {
	args := m.Called(ctx, userID, bookingID, req)
	return args.Error(0)
}

func (m *MockBookingService) GetBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.GetBookingByIDRow, error) {
	args := m.Called(ctx, userID, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetBookingByIDRow), args.Error(1)
}

func (m *MockBookingService) GetBookings(ctx context.Context, userID uuid.UUID) ([]sqlc.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.Booking), args.Error(1)
}

func (m *MockBookingService) CheckBookingTime(ctx context.Context, serviceID uuid.UUID, bookingTime time.Time) (bool, error) {
	args := m.Called(ctx, serviceID, bookingTime)
	return args.Bool(0), args.Error(1)
}

func (m *MockBookingService) ServiceExists(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	args := m.Called(ctx, serviceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBookingService) CheckBookingOwnerships(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, performerID uuid.UUID) bool {
	args := m.Called(ctx, userID, clientID, performerID)
	return args.Bool(0)
}

func (m *MockBookingService) DeleteBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error {
	args := m.Called(ctx, userID, bookingID)
	return args.Error(0)
}

var (
	testServiceID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testDiscountID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testPerformerID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testClientID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testBookingID   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

func setupTestHandler(t *testing.T) (*MockBookingService, *BookingHandler) {
	mockService := new(MockBookingService)
	logger := setupTestLogger()
	handler := NewBookingHandler(mockService, logger)
	return mockService, handler
}

func setupTestLogger() *slog.Logger {
	return slog.Default()
}

func setupContextWithUserID(userID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", userID.String())
	return ctx
}

func TestCreateBooking_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	reqBody := dto.CreateBookingRequest{
		ServiceID:   testServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
		DiscountID:  testDiscountID,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req = req.WithContext(setupContextWithUserID(testClientID))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	mockService.On("CreateBooking", mock.Anything, testClientID, mock.Anything).Return(testBookingID, nil)

	handler.CreateBooking()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CreateBookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, testBookingID, response.ID)

	mockService.AssertExpectations(t)
}

func TestCreateBooking_InvalidUserID(t *testing.T) {
	_, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.CreateBooking()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBooking_InvalidRequestBody(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(setupContextWithUserID(testClientID))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateBooking()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "CreateBooking")
}

func TestCreateBooking_ServiceError(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	reqBody := dto.CreateBookingRequest{
		ServiceID:   testServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req = req.WithContext(setupContextWithUserID(testClientID))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.On("CreateBooking", mock.Anything, testClientID, mock.Anything).Return(uuid.Nil, apierrors.ErrNotFound)

	handler.CreateBooking()(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCancelBooking_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	booking := &sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/"+testBookingID.String(), nil)
	req = req.WithContext(setupContextWithUserID(testClientID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("GetBooking", mock.Anything, testClientID, testBookingID).Return(booking, nil)
	mockService.On("CancelBooking", mock.Anything, testClientID, booking).Return(nil)

	handler.CancelBooking()(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestCancelBooking_InvalidBookingID(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/invalid-uuid", nil)
	req = req.WithContext(setupContextWithUserID(testClientID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.CancelBooking()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetBooking")
}

func TestCancelBooking_BookingNotFound(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/"+testBookingID.String(), nil)
	req = req.WithContext(setupContextWithUserID(testClientID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("GetBooking", mock.Anything, testClientID, testBookingID).Return(nil, apierrors.ErrNotFound)

	handler.CancelBooking()(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestSubmitBooking_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	booking := &sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
		Status:      "pending",
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/submit/"+testBookingID.String(), nil)
	req = req.WithContext(setupContextWithUserID(testPerformerID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("GetBooking", mock.Anything, testPerformerID, testBookingID).Return(booking, nil)
	mockService.On("SubmitBooking", mock.Anything, testPerformerID, booking).Return(nil)

	handler.SubmitBooking()(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestSubmitBooking_CancelledBooking(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	booking := &sqlc.GetBookingByIDRow{
		ID:          testBookingID,
		ClientID:    testClientID,
		PerformerID: testPerformerID,
		Status:      "cancelled",
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/submit/"+testBookingID.String(), nil)
	req = req.WithContext(setupContextWithUserID(testPerformerID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("GetBooking", mock.Anything, testPerformerID, testBookingID).Return(booking, nil)

	handler.SubmitBooking()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "SubmitBooking")
}

func TestGetBooking_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	expectedBooking := &sqlc.GetBookingByIDRow{
		ID:           testBookingID,
		ClientID:     testClientID,
		PerformerID:  testPerformerID,
		ServiceID:    testServiceID,
		ServiceTitle: "Test Service",
		BasePrice:    1000,
		FinalPrice:   800,
		Status:       "pending",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+testBookingID.String(), nil)
	req = req.WithContext(setupContextWithUserID(testClientID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("GetBooking", mock.Anything, testClientID, testBookingID).Return(expectedBooking, nil)

	handler.GetBooking()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.BookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, testBookingID.String(), response.ID)
	assert.Equal(t, "Test Service", response.ServiceTitle)

	mockService.AssertExpectations(t)
}

func TestGetBookings_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	expectedBookings := []sqlc.Booking{
		{ID: testBookingID, ClientID: testClientID},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req = req.WithContext(setupContextWithUserID(testClientID))
	w := httptest.NewRecorder()

	mockService.On("GetBookings", mock.Anything, testClientID).Return(expectedBookings, nil)

	handler.GetBookings()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responses []dto.BookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &responses)
	require.NoError(t, err)
	assert.Len(t, responses, 1)

	mockService.AssertExpectations(t)
}

func TestGetBookings_ServiceError(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req = req.WithContext(setupContextWithUserID(testClientID))
	w := httptest.NewRecorder()

	mockService.On("GetBookings", mock.Anything, testClientID).Return(nil, apierrors.ErrForbidden)

	handler.GetBookings()(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteBooking_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings/"+testBookingID.String(), nil)
	req = req.WithContext(setupContextWithUserID(testClientID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("DeleteBooking", mock.Anything, testClientID, testBookingID).Return(nil)

	handler.DeleteBooking()(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteBooking_InvalidID(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings/invalid-uuid", nil)
	req = req.WithContext(setupContextWithUserID(testClientID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.DeleteBooking()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "DeleteBooking")
}

func TestPatchBooking_Success(t *testing.T) {
	mockService, handler := setupTestHandler(t)

	newTime := time.Now().Add(48 * time.Hour)
	reqBody := dto.PatchBookingRequest{
		BookingTime: newTime,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/"+testBookingID.String(), bytes.NewReader(bodyBytes))
	req = req.WithContext(setupContextWithUserID(testClientID))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testBookingID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	mockService.On("UpdateBooking", mock.Anything, testClientID, testBookingID, mock.Anything).Return(nil)

	handler.PatchBooking()(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandleError(t *testing.T) {
	_, handler := setupTestHandler(t)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"ErrNotFound", apierrors.ErrNotFound, http.StatusNotFound},
		{"ErrForbidden", apierrors.ErrForbidden, http.StatusForbidden},
		{"ErrInvalidInput", apierrors.ErrInvalidInput, http.StatusBadRequest},
		{"ErrTimeBusy", apierrors.ErrTimeBusy, http.StatusConflict},
		{"ErrAlreadyDone", apierrors.ErrAlreadyDone, http.StatusConflict},
		{"Unknown error", errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			handler.writeError(w, r, tt.err, "")

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
