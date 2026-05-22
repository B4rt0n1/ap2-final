package booking

import (
	"encoding/json"
	"errors"
	"net/http"

	bookingv1 "github.com/B4rt0n1/final_proto/gen/go/booking/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler translates Booking HTTP routes into Booking Service gRPC calls.
type Handler struct {
	client bookingv1.BookingServiceClient
}

// Register adds Booking HTTP routes to the API Gateway mux.
func Register(mux *http.ServeMux, client bookingv1.BookingServiceClient) {
	handler := &Handler{client: client}

	mux.HandleFunc("POST /api/bookings", handler.createBooking)
	mux.HandleFunc("GET /api/bookings/{bookingID}", handler.getBooking)
	mux.HandleFunc("GET /api/users/{userID}/bookings", handler.listUserBookings)
	mux.HandleFunc("POST /api/bookings/{bookingID}/confirm", handler.confirmBooking)
	mux.HandleFunc("POST /api/bookings/{bookingID}/cancel", handler.cancelBooking)
}

type createBookingRequest struct {
	UserID    string `json:"user_id"`
	CarID     string `json:"car_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	var request createBookingRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.CreateBooking(r.Context(), &bookingv1.CreateBookingRequest{
		UserId:    request.UserID,
		CarId:     request.CarID,
		StartDate: request.StartDate,
		EndDate:   request.EndDate,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.GetBookingById(r.Context(), &bookingv1.GetBookingByIdRequest{
		BookingId: r.PathValue("bookingID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) listUserBookings(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.ListUserBookings(r.Context(), &bookingv1.ListUserBookingsRequest{
		UserId: r.PathValue("userID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) confirmBooking(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.ConfirmBooking(r.Context(), &bookingv1.ConfirmBookingRequest{
		BookingId: r.PathValue("bookingID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.CancelBooking(r.Context(), &bookingv1.CancelBookingRequest{
		BookingId: r.PathValue("bookingID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func decodeJSON(r *http.Request, value any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return errors.New("invalid JSON request body")
	}

	return nil
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, errors.New(st.Message()))
	case codes.NotFound:
		writeError(w, http.StatusNotFound, errors.New(st.Message()))
	case codes.FailedPrecondition:
		writeError(w, http.StatusConflict, errors.New(st.Message()))
	case codes.Unavailable:
		writeError(w, http.StatusServiceUnavailable, errors.New(st.Message()))
	default:
		writeError(w, http.StatusInternalServerError, errors.New(st.Message()))
	}
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, errorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
