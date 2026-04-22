//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== CREATE BOOKING TESTS ====================

func TestCreateBooking_Success(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	requestBody := dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
		DiscountID:  TestDiscountID,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+clientToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CreateBookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, response.ID)

	// Cleanup
	CleanupTestData(t)
}

func TestCreateBooking_NoAuth(t *testing.T) {
	// Arrange
	requestBody := dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateBooking_InvalidToken(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateBooking_InvalidRequestBody(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Authorization", "Bearer "+clientToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBooking_ForbiddenRole(t *testing.T) {
	// Arrange
	performerToken := GenerateTestToken(TestPerformerID, "performer")
	requestBody := dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+performerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateBooking_ServiceNotFound(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	requestBody := dto.CreateBookingRequest{
		ServiceID:   uuid.New(),
		BookingTime: time.Now().Add(24 * time.Hour),
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+clientToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateBooking_TimeBusy(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingTime := time.Now().Add(24 * time.Hour)

	// Создаем первое бронирование
	firstBooking := dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: bookingTime,
	}
	createBooking(t, clientToken, firstBooking)

	// Пытаемся создать второе на то же время
	secondBooking := dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: bookingTime,
	}
	bodyBytes, _ := json.Marshal(secondBooking)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+clientToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	// Cleanup
	CleanupTestData(t)
}

func TestCreateBooking_InvalidDiscount(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	requestBody := dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
		DiscountID:  uuid.New(), // Несуществующая скидка
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+clientToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==================== CANCEL BOOKING TESTS ====================

func TestCancelBooking_Success(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Cleanup
	CleanupTestData(t)
}

func TestCancelBooking_NotFound(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelBooking_Forbidden(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	otherClientToken := GenerateTestToken(uuid.New(), "client")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+otherClientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Cleanup
	CleanupTestData(t)
}

func TestCancelBooking_AdminCanCancelAny(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	adminToken := GenerateTestToken(TestAdminID, "admin")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Cleanup
	CleanupTestData(t)
}

// ==================== SUBMIT BOOKING TESTS ====================

func TestSubmitBooking_Success(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	performerToken := GenerateTestToken(TestPerformerID, "performer")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/submit/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+performerToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Cleanup
	CleanupTestData(t)
}

func TestSubmitBooking_NotPerformer(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	otherClientToken := GenerateTestToken(uuid.New(), "client")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/submit/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+otherClientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Cleanup
	CleanupTestData(t)
}

func TestSubmitBooking_AlreadySubmitted(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	performerToken := GenerateTestToken(TestPerformerID, "performer")

	// Первое подтверждение
	req1 := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/submit/"+bookingID.String(), nil)
	req1.Header.Set("Authorization", "Bearer "+performerToken)
	w1 := httptest.NewRecorder()
	testRouter.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusNoContent, w1.Code)

	// Второе подтверждение
	req2 := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/submit/"+bookingID.String(), nil)
	req2.Header.Set("Authorization", "Bearer "+performerToken)
	w2 := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w2, req2)

	// Assert
	assert.Equal(t, http.StatusConflict, w2.Code)

	// Cleanup
	CleanupTestData(t)
}

// ==================== PATCH BOOKING TESTS ====================

func TestPatchBooking_UpdateTimeByClient(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	newTime := time.Now().Add(48 * time.Hour)
	patchBody := dto.PatchBookingRequest{
		BookingTime: newTime,
	}
	bodyBytes, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/"+bookingID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+clientToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Cleanup
	CleanupTestData(t)
}

// ==================== GET BOOKING TESTS ====================

func TestGetBooking_Success(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.BookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, bookingID.String(), response.ID)

	// Cleanup
	CleanupTestData(t)
}

func TestGetBooking_NotFound(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==================== GET BOOKINGS LIST TESTS ====================

func TestGetBookings_AsClient(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")

	// Создаем несколько бронирований
	for i := 0; i < 3; i++ {
		createBooking(t, clientToken, dto.CreateBookingRequest{
			ServiceID:   TestServiceID,
			BookingTime: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var responses []dto.BookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &responses)
	require.NoError(t, err)
	assert.Len(t, responses, 3)

	// Cleanup
	CleanupTestData(t)
}

func TestGetBookings_Unauthorized(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== DELETE BOOKING TESTS ====================

func TestDeleteBooking_Success(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Проверяем, что бронирование удалено
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+bookingID.String(), nil)
	getReq.Header.Set("Authorization", "Bearer "+clientToken)
	getW := httptest.NewRecorder()
	testRouter.ServeHTTP(getW, getReq)
	assert.Equal(t, http.StatusNotFound, getW.Code)

	// Cleanup
	CleanupTestData(t)
}

func TestDeleteBooking_NotFound(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteBooking_Forbidden(t *testing.T) {
	// Arrange
	clientToken := GenerateTestToken(TestClientID, "client")
	bookingID := createBooking(t, clientToken, dto.CreateBookingRequest{
		ServiceID:   TestServiceID,
		BookingTime: time.Now().Add(24 * time.Hour),
	})

	otherClientToken := GenerateTestToken(uuid.New(), "client")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings/"+bookingID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+otherClientToken)
	w := httptest.NewRecorder()

	// Act
	testRouter.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Cleanup
	CleanupTestData(t)
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

func createBooking(t *testing.T, token string, req dto.CreateBookingRequest) uuid.UUID {
	bodyBytes, _ := json.Marshal(req)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	testRouter.ServeHTTP(response, request)

	// Добавим отладочный вывод
	if response.Code != http.StatusOK {
		t.Logf("Create booking failed with status %d: %s", response.Code, response.Body.String())
	}

	var createResponse dto.CreateBookingResponse
	err := json.Unmarshal(response.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	return createResponse.ID
}
