package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	pb "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GatewayHandler struct {
	client pb.CarInventoryServiceClient
}

func NewGatewayHandler(conn *grpc.ClientConn) *GatewayHandler {
	return &GatewayHandler{
		client: pb.NewCarInventoryServiceClient(conn),
	}
}

// Helper: write JSON error response
func writeError(w http.ResponseWriter, err error, defaultStatus int) {
	st, ok := status.FromError(err)
	code := defaultStatus
	msg := err.Error()
	if ok {
		// Map gRPC codes to HTTP status codes
		switch st.Code() {
		case 5: // NotFound
			code = http.StatusNotFound
		case 3: // InvalidArgument
			code = http.StatusBadRequest
		case 13: // Internal
			code = http.StatusInternalServerError
		default:
			code = http.StatusInternalServerError
		}
		msg = st.Message()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// 1. POST /cars
func (h *GatewayHandler) AddCar(w http.ResponseWriter, r *http.Request) {
	var req pb.AddCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.AddCar(ctx, &req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// 2. PUT /cars/{car_id}
func (h *GatewayHandler) UpdateCar(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id") // or use gorilla/mux: vars["car_id"]
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	var req pb.UpdateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	req.CarId = carID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.UpdateCar(ctx, &req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 3. DELETE /cars/{car_id}
func (h *GatewayHandler) DeleteCar(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	req := &pb.DeleteCarRequest{CarId: carID}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.DeleteCar(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 4. GET /cars/{car_id}
func (h *GatewayHandler) GetCarById(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	req := &pb.GetCarByIdRequest{CarId: carID}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GetCarById(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 5. GET /cars
func (h *GatewayHandler) ListCars(w http.ResponseWriter, r *http.Request) {
	req := &pb.ListCarsRequest{}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.ListCars(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 6. GET /cars/search
func (h *GatewayHandler) SearchCars(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	category := query.Get("category")
	minPrice, _ := strconv.ParseFloat(query.Get("min_price"), 32)
	maxPrice, _ := strconv.ParseFloat(query.Get("max_price"), 32)
	location := query.Get("location")

	req := &pb.SearchCarsRequest{
		Category: category,
		MinPrice: float64(minPrice),
		MaxPrice: float64(maxPrice),
		Location: location,
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.SearchCars(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 7. GET /cars/{car_id}/availability
func (h *GatewayHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if carID == "" || startDate == "" || endDate == "" {
		writeError(w, status.Error(3, "missing car_id, start_date or end_date"), http.StatusBadRequest)
		return
	}
	req := &pb.CheckAvailabilityRequest{
		CarId:     carID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.CheckAvailability(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 8. PATCH /cars/{car_id}/availability
func (h *GatewayHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	var req pb.UpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	req.CarId = carID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.UpdateAvailability(ctx, &req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 9. GET /cars/{car_id}/pricing
func (h *GatewayHandler) GetCarPricing(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	req := &pb.GetCarPricingRequest{CarId: carID}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GetCarPricing(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 10. PUT /cars/{car_id}/pricing
func (h *GatewayHandler) SetDynamicPricing(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	var req pb.SetDynamicPricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	req.CarId = carID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.SetDynamicPricing(ctx, &req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 11. POST /cars/{car_id}/images
func (h *GatewayHandler) UploadCarImages(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	var req pb.UploadCarImagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	req.CarId = carID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.UploadCarImages(ctx, &req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 12. GET /cars/{car_id}/reviews
func (h *GatewayHandler) GetCarReviews(w http.ResponseWriter, r *http.Request) {
	carID := r.URL.Query().Get("car_id")
	if carID == "" {
		writeError(w, status.Error(3, "missing car_id"), http.StatusBadRequest)
		return
	}
	req := &pb.GetCarReviewsRequest{CarId: carID}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GetCarReviews(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
