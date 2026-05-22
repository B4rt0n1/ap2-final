package usecase

import (
	"context"
	"testing"

	"inventory-service/internal/domain"
)

// Minimal mock implementation for the repository
type mockCarRepository struct {
	domain.CarRepository
	createFunc func(car *domain.Car) error
}

func (m *mockCarRepository) Create(ctx context.Context, car *domain.Car) error {
	return m.createFunc(car)
}

func TestAddCar_ValidatesPrice(t *testing.T) {
	repo := &mockCarRepository{}
	uc := NewCarUsecase(repo, nil, nil) // Leaving cache and publisher nil for this unit test
	invalidCar := &domain.Car{ID: "1", Brand: "Tesla", PricePerDay: -50.0}

	// Act
	err := uc.AddCar(context.Background(), invalidCar)

	// Assert
	if err == nil {
		t.Error("expected an error for negative car pricing, but got nil")
	}
	if err.Error() != "price per day must be strictly positive" {
		t.Errorf("unexpected error message: %v", err)
	}
}
