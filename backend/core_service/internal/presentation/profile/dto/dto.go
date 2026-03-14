package dto

type PatchProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PatchProfileResponse struct {
	Status string `json:"status"`
}
