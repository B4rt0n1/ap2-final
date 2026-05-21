package domain

import "errors"

var (
	ErrBookingNotFound         = errors.New("booking not found")
	ErrInvalidBookingID        = errors.New("booking id is required")
	ErrInvalidUserID           = errors.New("user id is required")
	ErrInvalidCarID            = errors.New("car id is required")
	ErrInvalidRentalDates      = errors.New("rental end date must be after start date")
	ErrInvalidTotalPrice       = errors.New("booking total price cannot be negative")
	ErrInvalidBookingStatus    = errors.New("invalid booking status")
	ErrInvalidStatusTransition = errors.New("invalid booking status transition")
)
