package repository

import (
	"context"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
)

// BookingStore describes the persistence operations used by booking flows.
type BookingStore interface {
	Create(ctx context.Context, booking domain.Booking) (domain.Booking, error)
	GetByID(ctx context.Context, bookingID string) (domain.Booking, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Booking, error)
	Update(ctx context.Context, booking domain.Booking) (domain.Booking, error)
}

// Transactor runs a booking flow against a single database transaction.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(context.Context, BookingStore) error) error
}
