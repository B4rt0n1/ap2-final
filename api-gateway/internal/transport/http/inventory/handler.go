package inventory

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	inventoryv1 "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client inventoryv1.CarInventoryServiceClient
}

func Register(mux *http.ServeMux, client inventoryv1.CarInventoryServiceClient) {
	handler := &Handler{client: client}

	mux.HandleFunc("POST /api/cars", handler.addCar)
	mux.HandleFunc("GET /api/cars", handler.listCars)
	mux.HandleFunc("GET /api/cars/search", handler.searchCars)
	mux.HandleFunc("GET /api/cars/{carID}", handler.getCar)
	mux.HandleFunc("PUT /api/cars/{carID}", handler.updateCar)
	mux.HandleFunc("DELETE /api/cars/{carID}", handler.deleteCar)
	mux.HandleFunc("GET /api/cars/{carID}/availability", handler.checkAvailability)
	mux.HandleFunc("PATCH /api/cars/{carID}/availability", handler.updateAvailability)
	mux.HandleFunc("GET /api/cars/{carID}/pricing", handler.getPricing)
	mux.HandleFunc("PUT /api/cars/{carID}/pricing", handler.setPricing)
	mux.HandleFunc("POST /api/cars/{carID}/images", handler.uploadImages)
	mux.HandleFunc("GET /api/cars/{carID}/reviews", handler.getReviews)
}

type addCarRequest struct {
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Year        int32   `json:"year"`
	Category    string  `json:"category"`
	PricePerDay float64 `json:"price_per_day"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type updateCarRequest struct {
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	PricePerDay float64 `json:"price_per_day"`
	Available   bool    `json:"available"`
}

type availabilityRequest struct {
	Available bool `json:"available"`
}

type pricingRequest struct {
	NewPrice float64 `json:"new_price"`
}

type imageRequest struct {
	ImageURLs []string `json:"image_urls"`
}

func (h *Handler) addCar(w http.ResponseWriter, r *http.Request) {
	var request addCarRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.AddCar(r.Context(), &inventoryv1.AddCarRequest{
		Brand:       request.Brand,
		Model:       request.Model,
		Year:        request.Year,
		Category:    request.Category,
		PricePerDay: request.PricePerDay,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) listCars(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.ListCars(r.Context(), &inventoryv1.ListCarsRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) searchCars(w http.ResponseWriter, r *http.Request) {
	minPrice, err := queryFloat(r, "min_price")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	maxPrice, err := queryFloat(r, "max_price")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.SearchCars(r.Context(), &inventoryv1.SearchCarsRequest{
		Category: r.URL.Query().Get("category"),
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Location: r.URL.Query().Get("location"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) getCar(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.GetCarById(r.Context(), &inventoryv1.GetCarByIdRequest{
		CarId: r.PathValue("carID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) updateCar(w http.ResponseWriter, r *http.Request) {
	var request updateCarRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.UpdateCar(r.Context(), &inventoryv1.UpdateCarRequest{
		CarId:       r.PathValue("carID"),
		Brand:       request.Brand,
		Model:       request.Model,
		PricePerDay: request.PricePerDay,
		Available:   request.Available,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) deleteCar(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.DeleteCar(r.Context(), &inventoryv1.DeleteCarRequest{CarId: r.PathValue("carID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) checkAvailability(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.CheckAvailability(r.Context(), &inventoryv1.CheckAvailabilityRequest{
		CarId:     r.PathValue("carID"),
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) updateAvailability(w http.ResponseWriter, r *http.Request) {
	var request availabilityRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.UpdateAvailability(r.Context(), &inventoryv1.UpdateAvailabilityRequest{
		CarId:     r.PathValue("carID"),
		Available: request.Available,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) getPricing(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.GetCarPricing(r.Context(), &inventoryv1.GetCarPricingRequest{CarId: r.PathValue("carID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) setPricing(w http.ResponseWriter, r *http.Request) {
	var request pricingRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.SetDynamicPricing(r.Context(), &inventoryv1.SetDynamicPricingRequest{
		CarId:    r.PathValue("carID"),
		NewPrice: request.NewPrice,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) uploadImages(w http.ResponseWriter, r *http.Request) {
	var request imageRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.client.UploadCarImages(r.Context(), &inventoryv1.UploadCarImagesRequest{
		CarId:     r.PathValue("carID"),
		ImageUrls: request.ImageURLs,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) getReviews(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.GetCarReviews(r.Context(), &inventoryv1.GetCarReviewsRequest{CarId: r.PathValue("carID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func queryFloat(r *http.Request, key string) (float64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, errors.New(key + " must be a number")
	}
	return value, nil
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
