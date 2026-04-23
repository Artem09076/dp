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

type BookingResponse struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	ServiceID     string    `json:"service_id"`
	ServiceTitle  string    `json:"service_title"`
	PerformerID   string    `json:"performer_id"`
	BasePrice     int32     `json:"base_price"`
	DiscountID    *string   `json:"discount_id,omitempty"`
	DiscountType  *string   `json:"discount_type,omitempty"`
	DiscountValue *int32    `json:"discount_value,omitempty"`
	FinalPrice    int32     `json:"final_price"`
	BookingTime   time.Time `json:"booking_time"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BookinListResponse struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	ServiceID     string    `json:"service_id"`
	BasePrice     int32     `json:"base_price"`
	DiscountID    *string   `json:"discount_id,omitempty"`
	DiscountType  *string   `json:"discount_type,omitempty"`
	DiscountValue *int32    `json:"discount_value,omitempty"`
	FinalPrice    int32     `json:"final_price"`
	BookingTime   time.Time `json:"booking_time"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
