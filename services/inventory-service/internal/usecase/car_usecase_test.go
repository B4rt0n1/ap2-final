package usecase

import (
	"context"
	"errors"
	"testing"

	"inventory-service/internal/domain"
)

type mockCarRepository struct {
	domain.CarRepository
	createFunc             func(car *domain.Car) error
	getByIDFunc            func(id string) (*domain.Car, error)
	searchFunc             func(category string, minPrice, maxPrice float64, location string) ([]*domain.Car, error)
	updateAvailabilityFunc func(id string, available bool) error
}

func (m *mockCarRepository) Create(ctx context.Context, car *domain.Car) error {
	return m.createFunc(car)
}

func (m *mockCarRepository) GetByID(_ context.Context, id string) (*domain.Car, error) {
	return m.getByIDFunc(id)
}

func (m *mockCarRepository) Search(_ context.Context, category string, minPrice, maxPrice float64, location string) ([]*domain.Car, error) {
	return m.searchFunc(category, minPrice, maxPrice, location)
}

func (m *mockCarRepository) UpdateAvailability(_ context.Context, id string, available bool) error {
	return m.updateAvailabilityFunc(id, available)
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

func TestGetCarByIDReturnsCachedCar(t *testing.T) {
	cache := &mockCarCache{
		getFunc: func(id string) (*domain.Car, error) {
			return &domain.Car{ID: id, Brand: "Tesla"}, nil
		},
	}
	repo := &mockCarRepository{
		getByIDFunc: func(id string) (*domain.Car, error) {
			t.Fatalf("repository should not be called on cache hit")
			return nil, nil
		},
	}
	uc := NewCarUsecase(repo, cache, nil)

	car, err := uc.GetCarByID(context.Background(), "car-1")
	if err != nil {
		t.Fatalf("GetCarByID() error = %v", err)
	}
	if car.ID != "car-1" || car.Brand != "Tesla" {
		t.Fatalf("GetCarByID() car = %#v", car)
	}
}

func TestGetCarByIDFallsBackToRepositoryOnCacheMiss(t *testing.T) {
	cache := &mockCarCache{
		getFunc: func(string) (*domain.Car, error) {
			return nil, errors.New("cache miss")
		},
		setFunc: func(*domain.Car) error {
			return nil
		},
	}
	repo := &mockCarRepository{
		getByIDFunc: func(id string) (*domain.Car, error) {
			return &domain.Car{ID: id, Brand: "Toyota"}, nil
		},
	}
	uc := NewCarUsecase(repo, cache, nil)

	car, err := uc.GetCarByID(context.Background(), "car-2")
	if err != nil {
		t.Fatalf("GetCarByID() error = %v", err)
	}
	if car.ID != "car-2" || car.Brand != "Toyota" {
		t.Fatalf("GetCarByID() car = %#v", car)
	}
}

func TestSearchCarsRejectsInvertedPriceRange(t *testing.T) {
	repo := &mockCarRepository{
		searchFunc: func(string, float64, float64, string) ([]*domain.Car, error) {
			t.Fatalf("repository should not be called for invalid price range")
			return nil, nil
		},
	}
	uc := NewCarUsecase(repo, nil, nil)

	if _, err := uc.SearchCars(context.Background(), "suv", 100, 50, "Almaty"); err == nil {
		t.Fatal("SearchCars() error = nil, want validation error")
	}
}

func TestUpdateAvailabilityPersistsAndInvalidatesCache(t *testing.T) {
	var updatedID string
	var updatedAvailability bool
	repo := &mockCarRepository{
		updateAvailabilityFunc: func(id string, available bool) error {
			updatedID = id
			updatedAvailability = available
			return nil
		},
	}
	cache := &mockCarCache{
		deleteFunc: func(id string) error {
			if id != "car-1" {
				t.Fatalf("cache delete id = %q, want car-1", id)
			}
			return nil
		},
	}
	publisher := &mockEventPublisher{}
	uc := NewCarUsecase(repo, cache, publisher)

	if err := uc.UpdateAvailability(context.Background(), "car-1", false); err != nil {
		t.Fatalf("UpdateAvailability() error = %v", err)
	}
	if updatedID != "car-1" || updatedAvailability {
		t.Fatalf("updated availability = id %q available %v", updatedID, updatedAvailability)
	}
}

func TestUploadCarImagesRejectsEmptyURLs(t *testing.T) {
	uc := NewCarUsecase(&mockCarRepository{}, nil, nil)

	if err := uc.UploadCarImages(context.Background(), "car-1", nil); err == nil {
		t.Fatal("UploadCarImages() error = nil, want validation error")
	}
}

type mockCarCache struct {
	getFunc    func(id string) (*domain.Car, error)
	setFunc    func(car *domain.Car) error
	deleteFunc func(id string) error
}

func (m *mockCarCache) Get(_ context.Context, id string) (*domain.Car, error) {
	return m.getFunc(id)
}

func (m *mockCarCache) Set(_ context.Context, car *domain.Car) error {
	if m.setFunc == nil {
		return nil
	}
	return m.setFunc(car)
}

func (m *mockCarCache) Delete(_ context.Context, id string) error {
	if m.deleteFunc == nil {
		return nil
	}
	return m.deleteFunc(id)
}

type mockEventPublisher struct{}

func (m *mockEventPublisher) PublishCarCreated(context.Context, *domain.Car) error {
	return nil
}

func (m *mockEventPublisher) PublishAvailabilityUpdated(context.Context, string, bool) error {
	return nil
}

func (m *mockEventPublisher) PublishPricingUpdated(context.Context, string, float64) error {
	return nil
}
