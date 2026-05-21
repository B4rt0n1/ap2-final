package domain

import (
	"fmt"
	"time"
)

// Status represents the booking lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

// Booking is the booking service's core rental reservation entity.
type Booking struct {
	ID         string
	UserID     string
	CarID      string
	StartDate  time.Time
	EndDate    time.Time
	Status     Status
	TotalPrice float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewBooking creates a pending reservation before persistence assigns metadata.
func NewBooking(userID, carID string, startDate, endDate time.Time, totalPrice float64) (Booking, error) {
	booking := Booking{
		UserID:     userID,
		CarID:      carID,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     StatusPending,
		TotalPrice: totalPrice,
	}

	if err := booking.ValidateForCreate(); err != nil {
		return Booking{}, err
	}

	return booking, nil
}

// ValidateForCreate checks the fields required before a new booking is stored.
func (b Booking) ValidateForCreate() error {
	if b.UserID == "" {
		return ErrInvalidUserID
	}

	if b.CarID == "" {
		return ErrInvalidCarID
	}

	return b.validateReservation()
}

// Validate checks a booking loaded from storage or prepared for a response.
func (b Booking) Validate() error {
	if b.ID == "" {
		return ErrInvalidBookingID
	}

	if err := b.ValidateForCreate(); err != nil {
		return err
	}

	return nil
}

// TransitionTo applies a valid booking lifecycle transition.
func (b *Booking) TransitionTo(next Status) error {
	if !next.IsValid() {
		return ErrInvalidBookingStatus
	}

	if b.Status == next {
		return nil
	}

	if !b.Status.canTransitionTo(next) {
		return fmt.Errorf("%w: %s to %s", ErrInvalidStatusTransition, b.Status, next)
	}

	b.Status = next
	return nil
}

func (b Booking) validateReservation() error {
	if !b.EndDate.After(b.StartDate) {
		return ErrInvalidRentalDates
	}

	if b.TotalPrice < 0 {
		return ErrInvalidTotalPrice
	}

	if !b.Status.IsValid() {
		return ErrInvalidBookingStatus
	}

	return nil
}

// IsValid reports whether the status is part of the booking lifecycle.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusCancelled, StatusActive, StatusCompleted:
		return true
	default:
		return false
	}
}

func (s Status) canTransitionTo(next Status) bool {
	switch s {
	case StatusPending:
		return next == StatusConfirmed || next == StatusCancelled
	case StatusConfirmed:
		return next == StatusActive || next == StatusCancelled
	case StatusActive:
		return next == StatusCompleted
	default:
		return false
	}
}
