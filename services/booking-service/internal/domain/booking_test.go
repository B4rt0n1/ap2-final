package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewBooking(t *testing.T) {
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(48 * time.Hour)

	booking, err := NewBooking("user-1", "car-1", startDate, endDate, 180)
	if err != nil {
		t.Fatalf("NewBooking() error = %v", err)
	}

	if booking.Status != StatusPending {
		t.Fatalf("NewBooking() status = %q, want %q", booking.Status, StatusPending)
	}
}

func TestNewBookingRejectsInvalidRentalDates(t *testing.T) {
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)

	_, err := NewBooking("user-1", "car-1", startDate, startDate, 180)
	if !errors.Is(err, ErrInvalidRentalDates) {
		t.Fatalf("NewBooking() error = %v, want %v", err, ErrInvalidRentalDates)
	}
}

func TestBookingTransitionTo(t *testing.T) {
	booking := Booking{Status: StatusPending}

	if err := booking.TransitionTo(StatusConfirmed); err != nil {
		t.Fatalf("TransitionTo() error = %v", err)
	}

	if booking.Status != StatusConfirmed {
		t.Fatalf("TransitionTo() status = %q, want %q", booking.Status, StatusConfirmed)
	}
}

func TestBookingTransitionToRejectsInvalidTransition(t *testing.T) {
	booking := Booking{Status: StatusCancelled}

	err := booking.TransitionTo(StatusActive)
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("TransitionTo() error = %v, want %v", err, ErrInvalidStatusTransition)
	}
}
