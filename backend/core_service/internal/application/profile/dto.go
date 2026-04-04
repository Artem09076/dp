package profile

type ProfileEventType string

const (
	ProfileVerificationStatusUpdatedSubmit ProfileEventType = "profile.verification_status_updated.1"
	ProfileVerificationStatusUpdatedReject ProfileEventType = "profile.verification_status_updated.2"
)

type ProfileEvent struct {
	Event              ProfileEventType `json:"event"`
	UserID             string           `json:"user_id"`
	Name               string           `json:"name"`
	Email              string           `json:"email"`
	VerificationStatus string           `json:"verification_status"`
}
