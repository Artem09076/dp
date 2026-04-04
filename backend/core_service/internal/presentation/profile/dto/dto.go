package dto

import "github.com/google/uuid"

type PatchProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PatchProfileResponse struct {
	Status string `json:"status"`
}

type UpdateVerificationStatusRequest struct {
	UserID             uuid.UUID `json:"user_id"`
	VerificationStatus string    `json:"verification_status"`
}
