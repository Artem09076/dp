package dto

import "github.com/google/uuid"

type CreateReviewRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	Rating    int32     `json:"rating"`
	Comment   string    `json:"comment"`
}

type ReviewResponse struct {
	ID        uuid.UUID `json:"id"`
	BookingID uuid.UUID `json:"booking_id"`
	Rating    int32     `json:"rating"`
	Comment   string    `json:"comment"`
}

type PatchReviewRequest struct {
	Rating  *int32  `json:"rating, omitempty"`
	Comment *string `json:"comment, omitempty"`
}
