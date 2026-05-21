package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/usecase"
	bookingv1 "github.com/B4rt0n1/final_proto/gen/go/booking/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateBookingMapsRequestAndResponse(t *testing.T) {
	service := &recordingBookingService{
		createBooking: booking("booking-1", domain.StatusPending),
	}
	handler := New(service)

	response, err := handler.CreateBooking(context.Background(), &bookingv1.CreateBookingRequest{
		UserId:    "user-1",
		CarId:     "car-1",
		StartDate: "2026-06-01",
		EndDate:   "2026-06-04T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	if service.createInput.UserID != "user-1" || service.createInput.CarID != "car-1" {
		t.Fatalf("CreateBooking() input = %#v", service.createInput)
	}
	if response.GetBooking().GetId() != "booking-1" {
		t.Fatalf("CreateBooking() response booking id = %q, want booking-1", response.GetBooking().GetId())
	}
	if response.GetBooking().GetStartDate() != "2026-06-01T00:00:00Z" {
		t.Fatalf("CreateBooking() response start date = %q", response.GetBooking().GetStartDate())
	}
}

func TestCreateBookingRejectsInvalidDate(t *testing.T) {
	handler := New(&recordingBookingService{})

	_, err := handler.CreateBooking(context.Background(), &bookingv1.CreateBookingRequest{
		UserId:    "user-1",
		CarId:     "car-1",
		StartDate: "tomorrow",
		EndDate:   "2026-06-04",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateBooking() status = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestGetBookingByIDMapsNotFound(t *testing.T) {
	handler := New(&recordingBookingService{getErr: domain.ErrBookingNotFound})

	_, err := handler.GetBookingById(context.Background(), &bookingv1.GetBookingByIdRequest{BookingId: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetBookingById() status = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestStartRentalReturnsMessage(t *testing.T) {
	handler := New(&recordingBookingService{startRental: booking("booking-1", domain.StatusActive)})

	response, err := handler.StartRental(context.Background(), &bookingv1.StartRentalRequest{BookingId: "booking-1"})
	if err != nil {
		t.Fatalf("StartRental() error = %v", err)
	}

	if response.GetMessage() != "rental started" {
		t.Fatalf("StartRental() message = %q, want rental started", response.GetMessage())
	}
}

type recordingBookingService struct {
	createInput   usecase.CreateBookingInput
	createBooking domain.Booking
	getErr        error
	startRental   domain.Booking
}

func (s *recordingBookingService) CreateBooking(_ context.Context, input usecase.CreateBookingInput) (domain.Booking, error) {
	s.createInput = input
	return s.createBooking, nil
}

func (s *recordingBookingService) GetBookingByID(context.Context, string) (domain.Booking, error) {
	if s.getErr != nil {
		return domain.Booking{}, s.getErr
	}

	return booking("booking-1", domain.StatusPending), nil
}

func (s *recordingBookingService) ListUserBookings(context.Context, string) ([]domain.Booking, error) {
	return []domain.Booking{booking("booking-1", domain.StatusPending)}, nil
}

func (s *recordingBookingService) CancelBooking(context.Context, string) (domain.Booking, error) {
	return booking("booking-1", domain.StatusCancelled), nil
}

func (s *recordingBookingService) UpdateBooking(context.Context, usecase.UpdateBookingInput) (domain.Booking, error) {
	return booking("booking-1", domain.StatusPending), nil
}

func (s *recordingBookingService) ConfirmBooking(context.Context, string) (domain.Booking, error) {
	return booking("booking-1", domain.StatusConfirmed), nil
}

func (s *recordingBookingService) StartRental(context.Context, string) (domain.Booking, error) {
	if s.startRental.ID == "" {
		return domain.Booking{}, errors.New("start rental not configured")
	}

	return s.startRental, nil
}

func (s *recordingBookingService) EndRental(context.Context, string) (domain.Booking, error) {
	return booking("booking-1", domain.StatusCompleted), nil
}

func (s *recordingBookingService) CalculateRentalCost(context.Context, string, int) (usecase.RentalCost, error) {
	return usecase.RentalCost{OriginalPrice: 120, DiscountedPrice: 120}, nil
}

func (s *recordingBookingService) ApplyDiscount(context.Context, string, string) (usecase.RentalCost, error) {
	return usecase.RentalCost{OriginalPrice: 120, DiscountedPrice: 108}, nil
}

func (s *recordingBookingService) GetBookingStatus(context.Context, string) (domain.Status, error) {
	return domain.StatusPending, nil
}

func (s *recordingBookingService) ReportIssue(context.Context, string, string) error {
	return nil
}

func booking(id string, bookingStatus domain.Status) domain.Booking {
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	return domain.Booking{
		ID:         id,
		UserID:     "user-1",
		CarID:      "car-1",
		StartDate:  startDate,
		EndDate:    startDate.Add(72 * time.Hour),
		Status:     bookingStatus,
		TotalPrice: 120,
	}
}
