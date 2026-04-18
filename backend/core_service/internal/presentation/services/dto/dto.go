package dto

import "github.com/google/uuid"

type CreateServiceRequest struct {
	PerformerID     uuid.UUID `json:"performer_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Price           int       `json:"price"`
	DurationMinutes int       `json:"duration_minutes"`
}

type PatchServiceRequest struct {
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	Price           *int    `json:"price,omitempty"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
}

type ServiceResponse struct {
	ID              string  `json:"id"`
	PerformerID     string  `json:"performer_id"`
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	Price           int64   `json:"price"`
	DurationMinutes int32   `json:"duration_minutes"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	AverageRating   float64 `json:"average_rating,omitempty"`
}
