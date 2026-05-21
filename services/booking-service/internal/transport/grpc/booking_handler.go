package grpcserver

import (
	"context"
	"errors"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/usecase"
	bookingv1 "github.com/B4rt0n1/final_proto/gen/go/booking/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const dateOnlyLayout = "2006-01-02"

// BookingService lists the booking use cases exposed through gRPC.
type BookingService interface {
	CreateBooking(context.Context, usecase.CreateBookingInput) (domain.Booking, error)
	GetBookingByID(context.Context, string) (domain.Booking, error)
	ListUserBookings(context.Context, string) ([]domain.Booking, error)
	CancelBooking(context.Context, string) (domain.Booking, error)
	UpdateBooking(context.Context, usecase.UpdateBookingInput) (domain.Booking, error)
	ConfirmBooking(context.Context, string) (domain.Booking, error)
	StartRental(context.Context, string) (domain.Booking, error)
	EndRental(context.Context, string) (domain.Booking, error)
	CalculateRentalCost(context.Context, string, int) (usecase.RentalCost, error)
	ApplyDiscount(context.Context, string, string) (usecase.RentalCost, error)
	GetBookingStatus(context.Context, string) (domain.Status, error)
	ReportIssue(context.Context, string, string) error
}

// Handler maps the generated booking gRPC contract to application use cases.
type Handler struct {
	bookingv1.UnimplementedBookingServiceServer

	service BookingService
}

// New creates a booking gRPC handler.
func New(service BookingService) *Handler {
	return &Handler{service: service}
}

// Register adds the booking handler to a gRPC service registrar.
func Register(registrar grpc.ServiceRegistrar, service BookingService) {
	bookingv1.RegisterBookingServiceServer(registrar, New(service))
}

func (h *Handler) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.BookingResponse, error) {
	startDate, err := parseDate(req.GetStartDate())
	if err != nil {
		return nil, grpcError(err)
	}

	endDate, err := parseDate(req.GetEndDate())
	if err != nil {
		return nil, grpcError(err)
	}

	booking, err := h.service.CreateBooking(ctx, usecase.CreateBookingInput{
		UserID:    req.GetUserId(),
		CarID:     req.GetCarId(),
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return bookingResponse(booking), nil
}

func (h *Handler) GetBookingById(ctx context.Context, req *bookingv1.GetBookingByIdRequest) (*bookingv1.BookingResponse, error) {
	booking, err := h.service.GetBookingByID(ctx, req.GetBookingId())
	if err != nil {
		return nil, grpcError(err)
	}

	return bookingResponse(booking), nil
}

func (h *Handler) ListUserBookings(ctx context.Context, req *bookingv1.ListUserBookingsRequest) (*bookingv1.BookingsResponse, error) {
	bookings, err := h.service.ListUserBookings(ctx, req.GetUserId())
	if err != nil {
		return nil, grpcError(err)
	}

	response := &bookingv1.BookingsResponse{
		Bookings: make([]*bookingv1.Booking, 0, len(bookings)),
	}
	for _, booking := range bookings {
		response.Bookings = append(response.Bookings, bookingMessage(booking))
	}

	return response, nil
}

func (h *Handler) CancelBooking(ctx context.Context, req *bookingv1.CancelBookingRequest) (*bookingv1.MessageResponse, error) {
	if _, err := h.service.CancelBooking(ctx, req.GetBookingId()); err != nil {
		return nil, grpcError(err)
	}

	return messageResponse("booking cancelled"), nil
}

func (h *Handler) UpdateBooking(ctx context.Context, req *bookingv1.UpdateBookingRequest) (*bookingv1.BookingResponse, error) {
	startDate, err := parseDate(req.GetStartDate())
	if err != nil {
		return nil, grpcError(err)
	}

	endDate, err := parseDate(req.GetEndDate())
	if err != nil {
		return nil, grpcError(err)
	}

	booking, err := h.service.UpdateBooking(ctx, usecase.UpdateBookingInput{
		BookingID: req.GetBookingId(),
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return bookingResponse(booking), nil
}

func (h *Handler) ConfirmBooking(ctx context.Context, req *bookingv1.ConfirmBookingRequest) (*bookingv1.BookingResponse, error) {
	booking, err := h.service.ConfirmBooking(ctx, req.GetBookingId())
	if err != nil {
		return nil, grpcError(err)
	}

	return bookingResponse(booking), nil
}

func (h *Handler) StartRental(ctx context.Context, req *bookingv1.StartRentalRequest) (*bookingv1.MessageResponse, error) {
	if _, err := h.service.StartRental(ctx, req.GetBookingId()); err != nil {
		return nil, grpcError(err)
	}

	return messageResponse("rental started"), nil
}

func (h *Handler) EndRental(ctx context.Context, req *bookingv1.EndRentalRequest) (*bookingv1.MessageResponse, error) {
	if _, err := h.service.EndRental(ctx, req.GetBookingId()); err != nil {
		return nil, grpcError(err)
	}

	return messageResponse("rental ended"), nil
}

func (h *Handler) CalculateRentalCost(ctx context.Context, req *bookingv1.CalculateRentalCostRequest) (*bookingv1.RentalCostResponse, error) {
	cost, err := h.service.CalculateRentalCost(ctx, req.GetCarId(), int(req.GetRentalDays()))
	if err != nil {
		return nil, grpcError(err)
	}

	return rentalCostResponse(cost), nil
}

func (h *Handler) ApplyDiscount(ctx context.Context, req *bookingv1.ApplyDiscountRequest) (*bookingv1.RentalCostResponse, error) {
	cost, err := h.service.ApplyDiscount(ctx, req.GetBookingId(), req.GetDiscountCode())
	if err != nil {
		return nil, grpcError(err)
	}

	return rentalCostResponse(cost), nil
}

func (h *Handler) GetBookingStatus(ctx context.Context, req *bookingv1.GetBookingStatusRequest) (*bookingv1.BookingStatusResponse, error) {
	bookingStatus, err := h.service.GetBookingStatus(ctx, req.GetBookingId())
	if err != nil {
		return nil, grpcError(err)
	}

	return &bookingv1.BookingStatusResponse{Status: string(bookingStatus)}, nil
}

func (h *Handler) ReportIssue(ctx context.Context, req *bookingv1.ReportIssueRequest) (*bookingv1.MessageResponse, error) {
	if err := h.service.ReportIssue(ctx, req.GetBookingId(), req.GetIssueDescription()); err != nil {
		return nil, grpcError(err)
	}

	return messageResponse("booking issue reported"), nil
}

func bookingResponse(booking domain.Booking) *bookingv1.BookingResponse {
	return &bookingv1.BookingResponse{Booking: bookingMessage(booking)}
}

func bookingMessage(booking domain.Booking) *bookingv1.Booking {
	return &bookingv1.Booking{
		Id:         booking.ID,
		UserId:     booking.UserID,
		CarId:      booking.CarID,
		StartDate:  booking.StartDate.UTC().Format(time.RFC3339),
		EndDate:    booking.EndDate.UTC().Format(time.RFC3339),
		Status:     string(booking.Status),
		TotalPrice: booking.TotalPrice,
	}
}

func messageResponse(message string) *bookingv1.MessageResponse {
	return &bookingv1.MessageResponse{Message: message}
}

func rentalCostResponse(cost usecase.RentalCost) *bookingv1.RentalCostResponse {
	return &bookingv1.RentalCostResponse{
		OriginalPrice:   cost.OriginalPrice,
		DiscountedPrice: cost.DiscountedPrice,
	}
}

func parseDate(raw string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}

	if parsed, err := time.Parse(dateOnlyLayout, raw); err == nil {
		return parsed, nil
	}

	return time.Time{}, domain.ErrInvalidRentalDates
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, domain.ErrBookingNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, usecase.ErrBookingNotEditable),
		errors.Is(err, usecase.ErrDiscountNotApplicable),
		errors.Is(err, usecase.ErrCarUnavailable):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrInvalidBookingID),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidCarID),
		errors.Is(err, domain.ErrInvalidRentalDates),
		errors.Is(err, domain.ErrInvalidTotalPrice),
		errors.Is(err, domain.ErrInvalidBookingStatus),
		errors.Is(err, usecase.ErrInvalidRentalDays),
		errors.Is(err, usecase.ErrInvalidDailyPrice),
		errors.Is(err, usecase.ErrInvalidDiscountCode),
		errors.Is(err, usecase.ErrInvalidIssue):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrPricingUnavailable),
		errors.Is(err, usecase.ErrUserValidationMissing),
		errors.Is(err, usecase.ErrCarValidationMissing):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
