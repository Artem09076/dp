package errors

import (
	"errors"
	"net/http"
)

// Domain errors
var (
	ErrNotFound              = errors.New("not found")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidInput          = errors.New("invalid input")
	ErrAlreadyExists         = errors.New("already exists")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrInvalidUserRole       = errors.New("invalid user role")
	ErrServiceNotFound       = errors.New("service not found")
	ErrDiscountInvalid       = errors.New("discount is invalid or expired")
	ErrDiscountUsageExceeded = errors.New("discount usage limit exceeded")
	ErrPerformerNotVerified  = errors.New("performer is not verified")
	ErrReviewAlreadyExists   = errors.New("review already exists for this booking")
	ErrInvalidReviewStatus   = errors.New("can only review completed or cancelled bookings")
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
func MapError(err error) *APIError {
	switch {
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
	case errors.Is(err, ErrAlreadyExists):
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "RESOURCE_ALREADY_EXISTS",
			Message:    "The resource already exists",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrInvalidUserRole):
		return &APIError{
			StatusCode: http.StatusForbidden,
			Code:       "INVALID_USER_ROLE",
			Message:    "Your user role doesn't allow this operation",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrServiceNotFound):
		return &APIError{
			StatusCode: http.StatusNotFound,
			Code:       "SERVICE_NOT_FOUND",
			Message:    "The requested service does not exist",
			Details:    "Please check the service ID and try again",
		}
	case errors.Is(err, ErrDiscountInvalid):
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "DISCOUNT_INVALID",
			Message:    "The discount is invalid or expired",
			Details:    "Please check the discount conditions",
		}
	case errors.Is(err, ErrDiscountUsageExceeded):
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "DISCOUNT_USAGE_EXCEEDED",
			Message:    "This discount has reached its usage limit",
			Details:    err.Error(),
		}
	case errors.Is(err, ErrPerformerNotVerified):
		return &APIError{
			StatusCode: http.StatusForbidden,
			Code:       "PERFORMER_NOT_VERIFIED",
			Message:    "Performer must be verified to create services",
			Details:    "Please complete verification process",
		}
	case errors.Is(err, ErrReviewAlreadyExists):
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "REVIEW_ALREADY_EXISTS",
			Message:    "A review already exists for this booking",
			Details:    "You can only create one review per booking",
		}
	case errors.Is(err, ErrInvalidReviewStatus):
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_REVIEW_STATUS",
			Message:    "Can only review completed or cancelled bookings",
			Details:    "Reviews are only allowed after booking is completed or cancelled",
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
