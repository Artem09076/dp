package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateBookingRequest struct {
	ServiceID   uuid.UUID `json:"service_id"`
	BookingTime time.Time `json:"booking_time"`
	DiscountID  uuid.UUID `json:"discount_id,omitempty"`
}

type CreateBookingResponse struct {
	ID uuid.UUID `json:"id"`
}

type CancelBookingResponse struct {
	Status string `json:"status"`
}

type PatchBookingRequest struct {
	BookingTime time.Time `json:"booking_time,omitempty"`
}
