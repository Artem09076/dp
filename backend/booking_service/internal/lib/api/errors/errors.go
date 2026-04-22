package errors

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidInput          = errors.New("invalid input")
	ErrTimeBusy              = errors.New("time slot already booked")
	ErrAlreadyDone           = errors.New("already processed")
	ErrServiceNotFound       = errors.New("service not found")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrInvalidUserRole       = errors.New("invalid user role")
	ErrDiscountInvalid       = errors.New("discount is invalid or expired")
	ErrDiscountUsageExceeded = errors.New("discount usage limit exceeded")
)

type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

// Error mapping with user-friendly messages
func MapError(err error) *APIError {
	// ВАЖНО: проверяем специфичные ошибки ПЕРВЫМИ!
	switch {
	case errors.Is(err, ErrTimeBusy):
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "TIME_SLOT_BUSY",
			Message:    "The requested time slot is already booked",
			Details:    "Please choose another time for your booking",
		}
	case errors.Is(err, ErrDiscountInvalid):
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "DISCOUNT_INVALID",
			Message:    "The discount code is invalid or expired",
			Details:    "Please check the discount conditions",
		}
	case errors.Is(err, ErrDiscountUsageExceeded):
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "DISCOUNT_USAGE_EXCEEDED",
			Message:    "This discount has reached its usage limit",
			Details:    "Please try another discount code",
		}
	case errors.Is(err, ErrInvalidUserRole):
		return &APIError{
			StatusCode: http.StatusForbidden,
			Code:       "INVALID_USER_ROLE",
			Message:    "Your user role doesn't allow this operation",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrAlreadyDone):
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "ALREADY_PROCESSED",
			Message:    "This booking has already been processed",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrServiceNotFound):
		return &APIError{
			StatusCode: http.StatusNotFound,
			Code:       "SERVICE_NOT_FOUND",
			Message:    "The requested service does not exist",
			Details:    "Please check the service ID and try again",
		}
	case errors.Is(err, ErrNotFound):
		return &APIError{
			StatusCode: http.StatusNotFound,
			Code:       "RESOURCE_NOT_FOUND",
			Message:    "The requested resource could not be found",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrForbidden):
		return &APIError{
			StatusCode: http.StatusForbidden,
			Code:       "ACCESS_FORBIDDEN",
			Message:    "You don't have permission to access this resource",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrUnauthorized):
		return &APIError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Authentication is required to access this resource",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrInvalidInput):
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_INPUT",
			Message:    "The request contains invalid data",
			Details:    err.Error(),
		}
	default:
		return &APIError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "An unexpected error occurred",
			Details:    err.Error(),
		}
	}
}

type ErrorResponse struct {
	Error *APIError `json:"error"`
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{
		Error: MapError(err),
	}
}
