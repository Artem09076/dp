package dto

import "github.com/google/uuid"

type CreateReviewRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	Rating    int32     `json:"rating"`
	Comment   string    `json:"comment"`
}

type ReviewResponse struct {
	ID        string  `json:"id"`
	BookingID string  `json:"booking_id"`
	Rating    int32   `json:"rating"`
	Comment   *string `json:"comment,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type PatchReviewRequest struct {
	Rating  *int32  `json:"rating, omitempty"`
	Comment *string `json:"comment, omitempty"`
}
