package booking

type BookingEventType string

const (
	BookingCreated   BookingEventType = "booking.created"
	BookingCancelled BookingEventType = "booking.cancelled"
	BookingSubmit    BookingEventType = "booking.submit"
)

type BookingEvent struct {
	Event     BookingEventType `json:"event"`
	BookingID string           `json:"booking_id"`
	Email     string           `json:"email"`
	Service   string           `json:"service"`
	Time      string           `json:"time"`
}
