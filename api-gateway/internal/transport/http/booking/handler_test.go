package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	bookingv1 "github.com/B4rt0n1/final_proto/gen/go/booking/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateBookingCallsGRPCAndReturnsCreated(t *testing.T) {
	client := &recordingClient{
		createResponse: &bookingv1.BookingResponse{
			Booking: &bookingv1.Booking{Id: "booking-1"},
		},
	}
	server := bookingMux(client)

	body := bytes.NewBufferString(`{"user_id":"user-1","car_id":"car-1","start_date":"2026-06-01","end_date":"2026-06-04"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/bookings", body)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("CreateBooking status = %d, want %d", response.Code, http.StatusCreated)
	}
	if client.createRequest.GetUserId() != "user-1" {
		t.Fatalf("CreateBooking gRPC user id = %q, want user-1", client.createRequest.GetUserId())
	}
}

func TestGetBookingMapsGRPCNotFound(t *testing.T) {
	server := bookingMux(&recordingClient{getErr: status.Error(codes.NotFound, "booking not found")})

	request := httptest.NewRequest(http.MethodGet, "/api/bookings/missing", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("GetBooking status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestConfirmBookingUsesPathValue(t *testing.T) {
	client := &recordingClient{
		confirmResponse: &bookingv1.BookingResponse{Booking: &bookingv1.Booking{Id: "booking-1"}},
	}
	server := bookingMux(client)

	request := httptest.NewRequest(http.MethodPost, "/api/bookings/booking-1/confirm", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("ConfirmBooking status = %d, want %d", response.Code, http.StatusOK)
	}
	if client.confirmRequest.GetBookingId() != "booking-1" {
		t.Fatalf("ConfirmBooking booking id = %q, want booking-1", client.confirmRequest.GetBookingId())
	}
}

func bookingMux(client bookingv1.BookingServiceClient) http.Handler {
	mux := http.NewServeMux()
	Register(mux, client)
	return mux
}

type recordingClient struct {
	createRequest   *bookingv1.CreateBookingRequest
	confirmRequest  *bookingv1.ConfirmBookingRequest
	createResponse  *bookingv1.BookingResponse
	confirmResponse *bookingv1.BookingResponse
	getErr          error
}

func (c *recordingClient) CreateBooking(_ context.Context, request *bookingv1.CreateBookingRequest, _ ...grpc.CallOption) (*bookingv1.BookingResponse, error) {
	c.createRequest = request
	return c.createResponse, nil
}

func (c *recordingClient) GetBookingById(context.Context, *bookingv1.GetBookingByIdRequest, ...grpc.CallOption) (*bookingv1.BookingResponse, error) {
	if c.getErr != nil {
		return nil, c.getErr
	}
	return &bookingv1.BookingResponse{Booking: &bookingv1.Booking{Id: "booking-1"}}, nil
}

func (c *recordingClient) ListUserBookings(context.Context, *bookingv1.ListUserBookingsRequest, ...grpc.CallOption) (*bookingv1.BookingsResponse, error) {
	return &bookingv1.BookingsResponse{}, nil
}

func (c *recordingClient) CancelBooking(context.Context, *bookingv1.CancelBookingRequest, ...grpc.CallOption) (*bookingv1.MessageResponse, error) {
	return &bookingv1.MessageResponse{Message: "booking cancelled"}, nil
}

func (c *recordingClient) UpdateBooking(context.Context, *bookingv1.UpdateBookingRequest, ...grpc.CallOption) (*bookingv1.BookingResponse, error) {
	return nil, nil
}

func (c *recordingClient) ConfirmBooking(_ context.Context, request *bookingv1.ConfirmBookingRequest, _ ...grpc.CallOption) (*bookingv1.BookingResponse, error) {
	c.confirmRequest = request
	return c.confirmResponse, nil
}

func (c *recordingClient) StartRental(context.Context, *bookingv1.StartRentalRequest, ...grpc.CallOption) (*bookingv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingClient) EndRental(context.Context, *bookingv1.EndRentalRequest, ...grpc.CallOption) (*bookingv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingClient) CalculateRentalCost(context.Context, *bookingv1.CalculateRentalCostRequest, ...grpc.CallOption) (*bookingv1.RentalCostResponse, error) {
	return nil, nil
}

func (c *recordingClient) ApplyDiscount(context.Context, *bookingv1.ApplyDiscountRequest, ...grpc.CallOption) (*bookingv1.RentalCostResponse, error) {
	return nil, nil
}

func (c *recordingClient) GetBookingStatus(context.Context, *bookingv1.GetBookingStatusRequest, ...grpc.CallOption) (*bookingv1.BookingStatusResponse, error) {
	return nil, nil
}

func (c *recordingClient) ReportIssue(context.Context, *bookingv1.ReportIssueRequest, ...grpc.CallOption) (*bookingv1.MessageResponse, error) {
	return nil, nil
}

func TestInvalidBodyReturnsJSONError(t *testing.T) {
	server := bookingMux(&recordingClient{})
	request := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewBufferString(`{bad`))
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	var got errorResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("Decode error response: %v", err)
	}
	if response.Code != http.StatusBadRequest || got.Error == "" {
		t.Fatalf("Invalid body status = %d, error = %q", response.Code, got.Error)
	}
}
