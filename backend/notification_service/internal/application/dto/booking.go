package dto

type BookingEventType string

const (
	BookingCreated   BookingEventType = "booking.created"
	BookingCancelled BookingEventType = "booking.cancelled"
	BookingSubmit    BookingEventType = "booking.submit"
	BookingUpdated1  BookingEventType = "booking.updated.1"
	BookingUpdated2  BookingEventType = "booking.updated.2"
)

type BookingEvent struct {
	Event     BookingEventType `json:"event"`
	BookingID string           `json:"booking_id"`
	Email     string           `json:"email"`
	Service   string           `json:"service"`
	Time      string           `json:"time"`
}
