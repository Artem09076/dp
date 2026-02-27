package dto

import "github.com/google/uuid"

type CreateServiceRequest struct {
	PerformerID     uuid.UUID `json:"performer_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Price           int       `json:"price"`
	DurationMinutes int       `json:"duration_minutes"`
}
