package usecase

import (
	"context"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
)

// EventType is the NATS subject for a booking lifecycle fact.
type EventType string

const (
	EventBookingCreated   EventType = "booking.created"
	EventBookingConfirmed EventType = "booking.confirmed"
	EventBookingCancelled EventType = "booking.cancelled"
)

// EventPublisher sends booking lifecycle events to a messaging boundary.
type EventPublisher interface {
	PublishBookingEvent(ctx context.Context, event BookingEvent) error
}

// BookingEvent is the queue payload other services can react to.
type BookingEvent struct {
	Type       EventType       `json:"type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Booking    BookingSnapshot `json:"booking"`
}

// BookingSnapshot avoids leaking domain methods into the queue payload.
type BookingSnapshot struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CarID      string    `json:"car_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
}

func newBookingEvent(eventType EventType, booking domain.Booking) BookingEvent {
	return BookingEvent{
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		Booking: BookingSnapshot{
			ID:         booking.ID,
			UserID:     booking.UserID,
			CarID:      booking.CarID,
			StartDate:  booking.StartDate,
			EndDate:    booking.EndDate,
			Status:     string(booking.Status),
			TotalPrice: booking.TotalPrice,
		},
	}
}
