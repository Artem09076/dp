package handlerlib

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid input")
	ErrTimeBusy        = errors.New("time slot already booked")
	ErrAlreadyDone     = errors.New("already processed")
	ErrServiceNotFound = errors.New("service not found")
)
