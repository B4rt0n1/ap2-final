package grpcclient

import (
	"context"
	"testing"
	"time"

	inventoryv1 "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
	"google.golang.org/grpc"
)

func TestUserClientChecksUserByID(t *testing.T) {
	client := NewUserClient(&recordingUserClient{
		response: &userv1.UserResponse{User: &userv1.User{Id: "user-1"}},
	})

	if err := client.EnsureUserExists(context.Background(), "user-1"); err != nil {
		t.Fatalf("EnsureUserExists() error = %v", err)
	}
}

func TestInventoryClientChecksAvailabilityAndFormatsDates(t *testing.T) {
	remote := &recordingInventoryClient{
		availability: &inventoryv1.AvailabilityResponse{Available: true},
	}
	client := NewInventoryClient(remote)
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)

	if err := client.EnsureCarAvailable(context.Background(), "car-1", startDate, startDate.Add(48*time.Hour)); err != nil {
		t.Fatalf("EnsureCarAvailable() error = %v", err)
	}

	if remote.availabilityRequest.GetStartDate() != "2026-06-01T00:00:00Z" {
		t.Fatalf("CheckAvailability() start date = %q", remote.availabilityRequest.GetStartDate())
	}
}

func TestInventoryClientGetsDailyPrice(t *testing.T) {
	client := NewInventoryClient(&recordingInventoryClient{
		pricing: &inventoryv1.PricingResponse{PricePerDay: 70},
	})

	price, err := client.DailyPrice(context.Background(), "car-1")
	if err != nil {
		t.Fatalf("DailyPrice() error = %v", err)
	}
	if price != 70 {
		t.Fatalf("DailyPrice() = %.2f, want 70.00", price)
	}
}

type recordingUserClient struct {
	response *userv1.UserResponse
	err      error
}

func (c *recordingUserClient) RegisterUser(context.Context, *userv1.RegisterUserRequest, ...grpc.CallOption) (*userv1.UserResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) LoginUser(context.Context, *userv1.LoginUserRequest, ...grpc.CallOption) (*userv1.AuthResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) LogoutUser(context.Context, *userv1.LogoutUserRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) GetUserById(context.Context, *userv1.GetUserByIdRequest, ...grpc.CallOption) (*userv1.UserResponse, error) {
	return c.response, c.err
}

func (c *recordingUserClient) UpdateUserProfile(context.Context, *userv1.UpdateUserProfileRequest, ...grpc.CallOption) (*userv1.UserResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) DeleteUser(context.Context, *userv1.DeleteUserRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) VerifyEmail(context.Context, *userv1.VerifyEmailRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) ResetPassword(context.Context, *userv1.ResetPasswordRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) ChangePassword(context.Context, *userv1.ChangePasswordRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) UploadDriverLicense(context.Context, *userv1.UploadDriverLicenseRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) GetUserRentalHistory(context.Context, *userv1.GetUserRentalHistoryRequest, ...grpc.CallOption) (*userv1.RentalHistoryResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) GetUserPaymentMethods(context.Context, *userv1.GetUserPaymentMethodsRequest, ...grpc.CallOption) (*userv1.PaymentMethodsResponse, error) {
	return nil, nil
}

type recordingInventoryClient struct {
	availabilityRequest *inventoryv1.CheckAvailabilityRequest
	availability        *inventoryv1.AvailabilityResponse
	pricing             *inventoryv1.PricingResponse
}

func (c *recordingInventoryClient) AddCar(context.Context, *inventoryv1.AddCarRequest, ...grpc.CallOption) (*inventoryv1.CarResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) UpdateCar(context.Context, *inventoryv1.UpdateCarRequest, ...grpc.CallOption) (*inventoryv1.CarResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) DeleteCar(context.Context, *inventoryv1.DeleteCarRequest, ...grpc.CallOption) (*inventoryv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) GetCarById(context.Context, *inventoryv1.GetCarByIdRequest, ...grpc.CallOption) (*inventoryv1.CarResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) ListCars(context.Context, *inventoryv1.ListCarsRequest, ...grpc.CallOption) (*inventoryv1.CarsResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) SearchCars(context.Context, *inventoryv1.SearchCarsRequest, ...grpc.CallOption) (*inventoryv1.CarsResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) CheckAvailability(_ context.Context, request *inventoryv1.CheckAvailabilityRequest, _ ...grpc.CallOption) (*inventoryv1.AvailabilityResponse, error) {
	c.availabilityRequest = request
	return c.availability, nil
}

func (c *recordingInventoryClient) UpdateAvailability(context.Context, *inventoryv1.UpdateAvailabilityRequest, ...grpc.CallOption) (*inventoryv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) GetCarPricing(context.Context, *inventoryv1.GetCarPricingRequest, ...grpc.CallOption) (*inventoryv1.PricingResponse, error) {
	return c.pricing, nil
}

func (c *recordingInventoryClient) SetDynamicPricing(context.Context, *inventoryv1.SetDynamicPricingRequest, ...grpc.CallOption) (*inventoryv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) UploadCarImages(context.Context, *inventoryv1.UploadCarImagesRequest, ...grpc.CallOption) (*inventoryv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingInventoryClient) GetCarReviews(context.Context, *inventoryv1.GetCarReviewsRequest, ...grpc.CallOption) (*inventoryv1.ReviewsResponse, error) {
	return nil, nil
}
