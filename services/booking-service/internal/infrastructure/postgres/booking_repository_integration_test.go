package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestBookingRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("BOOKING_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BOOKING_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db := openIntegrationDatabase(t, ctx, dsn)
	repo := NewBookingRepository(db)

	bookingID := "integration-booking-" + time.Now().UTC().Format("20060102150405.000000000")
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM bookings WHERE id = $1", bookingID); err != nil {
			t.Errorf("cleanup booking: %v", err)
		}
	})

	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	created, err := repo.Create(ctx, domain.Booking{
		ID:         bookingID,
		UserID:     "integration-user",
		CarID:      "integration-car",
		StartDate:  startDate,
		EndDate:    startDate.Add(72 * time.Hour),
		Status:     domain.StatusPending,
		TotalPrice: 225,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("Create() timestamps = created %v updated %v", created.CreatedAt, created.UpdatedAt)
	}

	got, err := repo.GetByID(ctx, bookingID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != bookingID || got.Status != domain.StatusPending {
		t.Fatalf("GetByID() booking = %#v", got)
	}

	if err := repo.WithinTransaction(ctx, func(ctx context.Context, store repository.BookingStore) error {
		booking, err := store.GetByID(ctx, bookingID)
		if err != nil {
			return err
		}

		booking.Status = domain.StatusConfirmed
		_, err = store.Update(ctx, booking)
		return err
	}); err != nil {
		t.Fatalf("WithinTransaction() error = %v", err)
	}

	updated, err := repo.GetByID(ctx, bookingID)
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}
	if updated.Status != domain.StatusConfirmed {
		t.Fatalf("updated status = %q, want %q", updated.Status, domain.StatusConfirmed)
	}
}

func openIntegrationDatabase(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() error = %v", err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext() error = %v", err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "000001_create_bookings_table.up.sql"))
	if err != nil {
		t.Fatalf("read booking migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil && !isRelationAlreadyExists(err) {
		t.Fatalf("apply booking migration: %v", err)
	}

	return db
}

func isRelationAlreadyExists(err error) bool {
	return err != nil && sqlState(err) == "42P07"
}

type sqlStateError interface {
	SQLState() string
}

func sqlState(err error) string {
	var value sqlStateError
	if errors.As(err, &value) {
		return value.SQLState()
	}

	return ""
}
