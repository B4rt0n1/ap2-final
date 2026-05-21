package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/repository"
)

var _ repository.BookingStore = (*BookingRepository)(nil)
var _ repository.Transactor = (*BookingRepository)(nil)

// BookingRepository stores booking entities in PostgreSQL.
type BookingRepository struct {
	db      *sql.DB
	queries bookingQueries
}

type bookingQueries interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type rowScanner interface {
	Scan(dest ...any) error
}

// NewBookingRepository creates a PostgreSQL-backed booking repository.
func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{
		db:      db,
		queries: db,
	}
}

// Create inserts a new booking and returns database-managed timestamps.
func (r *BookingRepository) Create(ctx context.Context, booking domain.Booking) (domain.Booking, error) {
	if err := booking.Validate(); err != nil {
		return domain.Booking{}, err
	}

	const query = `
		INSERT INTO bookings (
			id,
			user_id,
			car_id,
			start_date,
			end_date,
			status,
			total_price
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			user_id,
			car_id,
			start_date,
			end_date,
			status,
			total_price,
			created_at,
			updated_at`

	created, err := scanBooking(r.queries.QueryRowContext(
		ctx,
		query,
		booking.ID,
		booking.UserID,
		booking.CarID,
		booking.StartDate,
		booking.EndDate,
		booking.Status,
		booking.TotalPrice,
	))
	if err != nil {
		return domain.Booking{}, fmt.Errorf("create booking: %w", err)
	}

	return created, nil
}

// GetByID loads a booking by its public identifier.
func (r *BookingRepository) GetByID(ctx context.Context, bookingID string) (domain.Booking, error) {
	const query = `
		SELECT
			id,
			user_id,
			car_id,
			start_date,
			end_date,
			status,
			total_price,
			created_at,
			updated_at
		FROM bookings
		WHERE id = $1`

	booking, err := scanBooking(r.queries.QueryRowContext(ctx, query, bookingID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("get booking by id: %w", err)
	}

	return booking, nil
}

// ListByUserID returns a user's newest bookings first.
func (r *BookingRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Booking, error) {
	const query = `
		SELECT
			id,
			user_id,
			car_id,
			start_date,
			end_date,
			status,
			total_price,
			created_at,
			updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.queries.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user booking: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user bookings: %w", err)
	}

	return bookings, nil
}

// Update saves reservation dates, status, and total price changes.
func (r *BookingRepository) Update(ctx context.Context, booking domain.Booking) (domain.Booking, error) {
	if err := booking.Validate(); err != nil {
		return domain.Booking{}, err
	}

	const query = `
		UPDATE bookings
		SET
			start_date = $2,
			end_date = $3,
			status = $4,
			total_price = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			user_id,
			car_id,
			start_date,
			end_date,
			status,
			total_price,
			created_at,
			updated_at`

	updated, err := scanBooking(r.queries.QueryRowContext(
		ctx,
		query,
		booking.ID,
		booking.StartDate,
		booking.EndDate,
		booking.Status,
		booking.TotalPrice,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("update booking: %w", err)
	}

	return updated, nil
}

// WithinTransaction runs a booking flow on one SQL transaction.
func (r *BookingRepository) WithinTransaction(
	ctx context.Context,
	fn func(context.Context, repository.BookingStore) error,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin booking transaction: %w", err)
	}
	defer tx.Rollback()

	txRepository := &BookingRepository{
		db:      r.db,
		queries: tx,
	}

	if err := fn(ctx, txRepository); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit booking transaction: %w", err)
	}

	return nil
}

func scanBooking(scanner rowScanner) (domain.Booking, error) {
	var booking domain.Booking

	if err := scanner.Scan(
		&booking.ID,
		&booking.UserID,
		&booking.CarID,
		&booking.StartDate,
		&booking.EndDate,
		&booking.Status,
		&booking.TotalPrice,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	); err != nil {
		return domain.Booking{}, err
	}

	if err := booking.Validate(); err != nil {
		return domain.Booking{}, err
	}

	return booking, nil
}
