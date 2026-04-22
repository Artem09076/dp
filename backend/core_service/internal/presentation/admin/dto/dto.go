package dto

import (
	"time"
)

type UserResponse struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	INN                *string   `json:"inn,omitempty"`
	BusinessType       *string   `json:"business_type,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	ServicesCount      int64     `json:"services_count,omitempty"`
	TotalBookings      int64     `json:"total_bookings,omitempty"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"` // admin, client, performer
}

type VerifyPerformerRequest struct {
	Status          string `json:"status"` // verified, rejected
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type BatchVerifyRequest struct {
	UserIDs []string `json:"user_ids"`
	Status  string   `json:"status"` // verified, rejected
}

type ServiceResponse struct {
	ID              string    `json:"id"`
	PerformerID     string    `json:"performer_id"`
	PerformerName   string    `json:"performer_name,omitempty"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	Price           int64     `json:"price"`
	DurationMinutes int32     `json:"duration_minutes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type BookingResponse struct {
	ID          string    `json:"id"`
	ClientID    string    `json:"client_id"`
	ClientName  string    `json:"client_name,omitempty"`
	ServiceID   string    `json:"service_id"`
	ServiceName string    `json:"service_name,omitempty"`
	BasePrice   int64     `json:"base_price"`
	FinalPrice  int64     `json:"final_price"`
	BookingTime time.Time `json:"booking_time"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ReviewResponse struct {
	ID         string    `json:"id"`
	BookingID  string    `json:"booking_id"`
	Rating     int32     `json:"rating"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ClientName string    `json:"client_name,omitempty"`
	ServiceID  string    `json:"service_id,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
