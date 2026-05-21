package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/usecase"
	inventoryv1 "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
)

var _ usecase.UserValidator = (*UserClient)(nil)
var _ usecase.CarAvailabilityChecker = (*InventoryClient)(nil)
var _ usecase.PricingProvider = (*InventoryClient)(nil)

// UserClient validates booking users through User Service gRPC.
type UserClient struct {
	client userv1.UserServiceClient
}

// NewUserClient creates a User Service boundary for booking use cases.
func NewUserClient(client userv1.UserServiceClient) *UserClient {
	return &UserClient{client: client}
}

// EnsureUserExists returns an error when User Service cannot find the user.
func (c *UserClient) EnsureUserExists(ctx context.Context, userID string) error {
	response, err := c.client.GetUserById(ctx, &userv1.GetUserByIdRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("get booking user: %w", err)
	}
	if response.GetUser() == nil || response.GetUser().GetId() == "" {
		return errors.New("user service returned an empty user")
	}

	return nil
}

// InventoryClient validates availability and pricing through Inventory Service gRPC.
type InventoryClient struct {
	client inventoryv1.CarInventoryServiceClient
}

// NewInventoryClient creates an Inventory Service boundary for booking use cases.
func NewInventoryClient(client inventoryv1.CarInventoryServiceClient) *InventoryClient {
	return &InventoryClient{client: client}
}

// EnsureCarAvailable checks whether inventory can rent the car for the requested dates.
func (c *InventoryClient) EnsureCarAvailable(
	ctx context.Context,
	carID string,
	startDate time.Time,
	endDate time.Time,
) error {
	response, err := c.client.CheckAvailability(ctx, &inventoryv1.CheckAvailabilityRequest{
		CarId:     carID,
		StartDate: startDate.UTC().Format(time.RFC3339),
		EndDate:   endDate.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("check car availability: %w", err)
	}
	if !response.GetAvailable() {
		return usecase.ErrCarUnavailable
	}

	return nil
}

// DailyPrice returns the current Inventory Service price per rental day.
func (c *InventoryClient) DailyPrice(ctx context.Context, carID string) (float64, error) {
	response, err := c.client.GetCarPricing(ctx, &inventoryv1.GetCarPricingRequest{CarId: carID})
	if err != nil {
		return 0, fmt.Errorf("get car pricing: %w", err)
	}

	return response.GetPricePerDay(), nil
}
