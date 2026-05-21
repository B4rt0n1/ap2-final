package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/repository"
)

func TestCreateBookingUsesPricingAndTransaction(t *testing.T) {
	store := newMemoryBookingStore()
	tx := &memoryTransactor{store: store}
	service := New(Dependencies{
		Bookings:   store,
		Transactor: tx,
		IDs:        staticIDGenerator{id: "booking-1"},
		Pricing:    staticPricing{price: 45},
	})
	startDate, endDate := rentalWindow()

	booking, err := service.CreateBooking(context.Background(), CreateBookingInput{
		UserID:    "user-1",
		CarID:     "car-1",
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	if tx.calls != 1 {
		t.Fatalf("CreateBooking() transaction calls = %d, want 1", tx.calls)
	}
	if booking.ID != "booking-1" {
		t.Fatalf("CreateBooking() id = %q, want booking-1", booking.ID)
	}
	if booking.Status != domain.StatusPending {
		t.Fatalf("CreateBooking() status = %q, want %q", booking.Status, domain.StatusPending)
	}
	if booking.TotalPrice != 135 {
		t.Fatalf("CreateBooking() total price = %.2f, want 135.00", booking.TotalPrice)
	}
}

func TestConfirmBookingPersistsStatusTransition(t *testing.T) {
	store := newMemoryBookingStore(existingBooking("booking-1", domain.StatusPending))
	tx := &memoryTransactor{store: store}
	service := New(Dependencies{Bookings: store, Transactor: tx})

	booking, err := service.ConfirmBooking(context.Background(), "booking-1")
	if err != nil {
		t.Fatalf("ConfirmBooking() error = %v", err)
	}

	if booking.Status != domain.StatusConfirmed {
		t.Fatalf("ConfirmBooking() status = %q, want %q", booking.Status, domain.StatusConfirmed)
	}
}

func TestUpdateBookingRejectsConfirmedBooking(t *testing.T) {
	store := newMemoryBookingStore(existingBooking("booking-1", domain.StatusConfirmed))
	service := New(Dependencies{Bookings: store, Transactor: &memoryTransactor{store: store}})
	startDate, endDate := rentalWindow()

	_, err := service.UpdateBooking(context.Background(), UpdateBookingInput{
		BookingID: "booking-1",
		StartDate: startDate.Add(24 * time.Hour),
		EndDate:   endDate.Add(24 * time.Hour),
	})
	if !errors.Is(err, ErrBookingNotEditable) {
		t.Fatalf("UpdateBooking() error = %v, want %v", err, ErrBookingNotEditable)
	}
}

func TestApplyDiscountUpdatesPendingBooking(t *testing.T) {
	booking := existingBooking("booking-1", domain.StatusPending)
	booking.TotalPrice = 200
	store := newMemoryBookingStore(booking)
	service := New(Dependencies{Bookings: store, Transactor: &memoryTransactor{store: store}})

	cost, err := service.ApplyDiscount(context.Background(), "booking-1", "rent10")
	if err != nil {
		t.Fatalf("ApplyDiscount() error = %v", err)
	}

	if cost.OriginalPrice != 200 || cost.DiscountedPrice != 180 {
		t.Fatalf("ApplyDiscount() cost = %#v, want original 200 and discounted 180", cost)
	}
	if got := store.bookings["booking-1"].TotalPrice; got != 180 {
		t.Fatalf("ApplyDiscount() stored total price = %.2f, want 180.00", got)
	}
}

func TestReportIssueForwardsValidatedIssue(t *testing.T) {
	store := newMemoryBookingStore(existingBooking("booking-1", domain.StatusActive))
	reporter := &memoryIssueReporter{}
	service := New(Dependencies{
		Bookings:    store,
		IssueReport: reporter,
	})

	if err := service.ReportIssue(context.Background(), "booking-1", "flat tire"); err != nil {
		t.Fatalf("ReportIssue() error = %v", err)
	}

	if reporter.issue.BookingID != "booking-1" || reporter.issue.Description != "flat tire" {
		t.Fatalf("ReportIssue() issue = %#v", reporter.issue)
	}
}

type staticIDGenerator struct {
	id string
}

func (g staticIDGenerator) NewID() string {
	return g.id
}

type staticPricing struct {
	price float64
	err   error
}

func (p staticPricing) DailyPrice(context.Context, string) (float64, error) {
	return p.price, p.err
}

type memoryIssueReporter struct {
	issue IssueReport
}

func (r *memoryIssueReporter) ReportBookingIssue(_ context.Context, issue IssueReport) error {
	r.issue = issue
	return nil
}

type memoryTransactor struct {
	store *memoryBookingStore
	calls int
}

func (t *memoryTransactor) WithinTransaction(
	ctx context.Context,
	fn func(context.Context, repository.BookingStore) error,
) error {
	t.calls++
	return fn(ctx, t.store)
}

type memoryBookingStore struct {
	bookings map[string]domain.Booking
}

func newMemoryBookingStore(bookings ...domain.Booking) *memoryBookingStore {
	store := &memoryBookingStore{bookings: make(map[string]domain.Booking)}
	for _, booking := range bookings {
		store.bookings[booking.ID] = booking
	}

	return store
}

func (s *memoryBookingStore) Create(_ context.Context, booking domain.Booking) (domain.Booking, error) {
	s.bookings[booking.ID] = booking
	return booking, nil
}

func (s *memoryBookingStore) GetByID(_ context.Context, bookingID string) (domain.Booking, error) {
	booking, ok := s.bookings[bookingID]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}

	return booking, nil
}

func (s *memoryBookingStore) ListByUserID(_ context.Context, userID string) ([]domain.Booking, error) {
	bookings := make([]domain.Booking, 0)
	for _, booking := range s.bookings {
		if booking.UserID == userID {
			bookings = append(bookings, booking)
		}
	}

	return bookings, nil
}

func (s *memoryBookingStore) Update(_ context.Context, booking domain.Booking) (domain.Booking, error) {
	if _, ok := s.bookings[booking.ID]; !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}

	s.bookings[booking.ID] = booking
	return booking, nil
}

func existingBooking(id string, status domain.Status) domain.Booking {
	startDate, endDate := rentalWindow()
	return domain.Booking{
		ID:         id,
		UserID:     "user-1",
		CarID:      "car-1",
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     status,
		TotalPrice: 120,
	}
}

func rentalWindow() (time.Time, time.Time) {
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	return startDate, startDate.Add(72 * time.Hour)
}
