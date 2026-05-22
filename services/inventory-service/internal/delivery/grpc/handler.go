package grpc

import (
	"context"
	"inventory-service/internal/domain"

	pb "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CarInventoryHandler struct {
	pb.UnimplementedCarInventoryServiceServer
	carUsecase domain.CarUsecase
}

func NewCarInventoryHandler(u domain.CarUsecase) *CarInventoryHandler {
	return &CarInventoryHandler{carUsecase: u}
}

func mapDomainCarToProto(c *domain.Car) *pb.Car {
	return &pb.Car{
		Id:          c.ID,
		Brand:       c.Brand,
		Model:       c.Model,
		Year:        c.Year,
		Category:    c.Category,
		PricePerDay: c.PricePerDay,
		Available:   c.Available,
	}
}

// 1. AddCar
func (h *CarInventoryHandler) AddCar(ctx context.Context, req *pb.AddCarRequest) (*pb.CarResponse, error) {
	car := &domain.Car{
		// ID is intentionally left empty here; usecase will populate it
		Brand:       req.GetBrand(),
		Model:       req.GetModel(),
		Year:        req.GetYear(),
		Category:    req.GetCategory(),
		PricePerDay: req.GetPricePerDay(),
		Available:   true,
	}

	// When this executes, car.ID gets overwritten with a real UUID
	if err := h.carUsecase.AddCar(ctx, car); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to add car: %v", err)
	}

	// Returns the car payload containing the freshly generated UUID back to the client
	return &pb.CarResponse{
		Car: mapDomainCarToProto(car),
	}, nil
}

// 2. UpdateCar
func (h *CarInventoryHandler) UpdateCar(ctx context.Context, req *pb.UpdateCarRequest) (*pb.CarResponse, error) {
	car := &domain.Car{
		ID:          req.GetCarId(),
		Brand:       req.GetBrand(),
		Model:       req.GetModel(),
		PricePerDay: req.GetPricePerDay(),
		Available:   req.GetAvailable(),
	}
	if err := h.carUsecase.UpdateCar(ctx, car); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update car: %v", err)
	}
	return &pb.CarResponse{Car: mapDomainCarToProto(car)}, nil
}

// 3. DeleteCar
func (h *CarInventoryHandler) DeleteCar(ctx context.Context, req *pb.DeleteCarRequest) (*pb.MessageResponse, error) {
	if err := h.carUsecase.DeleteCar(ctx, req.GetCarId()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete car: %v", err)
	}
	return &pb.MessageResponse{Message: "Car deleted successfully"}, nil
}

// 4. GetCarById
func (h *CarInventoryHandler) GetCarById(ctx context.Context, req *pb.GetCarByIdRequest) (*pb.CarResponse, error) {
	car, err := h.carUsecase.GetCarByID(ctx, req.GetCarId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Car not found: %v", err)
	}

	return &pb.CarResponse{
		Car: mapDomainCarToProto(car),
	}, nil
}

// 5. ListCars
func (h *CarInventoryHandler) ListCars(ctx context.Context, req *pb.ListCarsRequest) (*pb.CarsResponse, error) {
	cars, err := h.carUsecase.ListCars(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list cars: %v", err)
	}
	var protoCars []*pb.Car
	for _, c := range cars {
		protoCars = append(protoCars, mapDomainCarToProto(c))
	}
	return &pb.CarsResponse{Cars: protoCars}, nil
}

// 6. SearchCars
func (h *CarInventoryHandler) SearchCars(ctx context.Context, req *pb.SearchCarsRequest) (*pb.CarsResponse, error) {
	cars, err := h.carUsecase.SearchCars(ctx, req.GetCategory(), req.GetMinPrice(), req.GetMaxPrice(), req.GetLocation())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Search failed: %v", err)
	}
	var protoCars []*pb.Car
	for _, c := range cars {
		protoCars = append(protoCars, mapDomainCarToProto(c))
	}
	return &pb.CarsResponse{Cars: protoCars}, nil
}

// 7. CheckAvailability
func (h *CarInventoryHandler) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.AvailabilityResponse, error) {
	available, err := h.carUsecase.CheckAvailability(ctx, req.GetCarId(), req.GetStartDate(), req.GetEndDate())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check availability: %v", err)
	}
	return &pb.AvailabilityResponse{Available: available}, nil
}

// 8. UpdateAvailability
func (h *CarInventoryHandler) UpdateAvailability(ctx context.Context, req *pb.UpdateAvailabilityRequest) (*pb.MessageResponse, error) {
	if err := h.carUsecase.UpdateAvailability(ctx, req.GetCarId(), req.GetAvailable()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update availability: %v", err)
	}
	return &pb.MessageResponse{Message: "Availability updated successfully"}, nil
}

// 9. GetCarPricing
func (h *CarInventoryHandler) GetCarPricing(ctx context.Context, req *pb.GetCarPricingRequest) (*pb.PricingResponse, error) {
	price, err := h.carUsecase.GetCarPricing(ctx, req.GetCarId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get pricing: %v", err)
	}
	return &pb.PricingResponse{PricePerDay: price}, nil
}

// 10. SetDynamicPricing
func (h *CarInventoryHandler) SetDynamicPricing(ctx context.Context, req *pb.SetDynamicPricingRequest) (*pb.MessageResponse, error) {
	if err := h.carUsecase.SetDynamicPricing(ctx, req.GetCarId(), req.GetNewPrice()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update pricing: %v", err)
	}
	return &pb.MessageResponse{Message: "Pricing updated successfully"}, nil
}

// 11. UploadCarImages
func (h *CarInventoryHandler) UploadCarImages(ctx context.Context, req *pb.UploadCarImagesRequest) (*pb.MessageResponse, error) {
	if err := h.carUsecase.UploadCarImages(ctx, req.GetCarId(), req.GetImageUrls()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to upload images: %v", err)
	}
	return &pb.MessageResponse{Message: "Images uploaded successfully"}, nil
}

// 12. GetCarReviews
func (h *CarInventoryHandler) GetCarReviews(ctx context.Context, req *pb.GetCarReviewsRequest) (*pb.ReviewsResponse, error) {
	reviews, err := h.carUsecase.GetCarReviews(ctx, req.GetCarId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch reviews: %v", err)
	}
	var protoReviews []*pb.Review
	for _, r := range reviews {
		protoReviews = append(protoReviews, &pb.Review{
			UserId:  r.UserID,
			Rating:  r.Rating,
			Comment: r.Comment,
		})
	}
	return &pb.ReviewsResponse{Reviews: protoReviews}, nil
}
