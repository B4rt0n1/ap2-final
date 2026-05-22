package usecase

import (
	"context"
	"errors"
	"inventory-service/internal/domain"
)

type carUsecase struct {
	carRepo domain.CarRepository
}

func NewCarUsecase(repo domain.CarRepository) domain.CarUsecase {
	return &carUsecase{carRepo: repo}
}

func (u *carUsecase) AddCar(ctx context.Context, car *domain.Car) error {
	if car.PricePerDay <= 0 {
		return errors.New("price per day must be strictly positive")
	}
	return u.carRepo.Create(ctx, car)
}

func (u *carUsecase) GetCarByID(ctx context.Context, id string) (*domain.Car, error) {
	return u.carRepo.GetByID(ctx, id)
}

func (u *carUsecase) UpdateCar(ctx context.Context, car *domain.Car) error {
	if car.ID == "" {
		return errors.New("car ID is required for update")
	}
	return u.carRepo.Update(ctx, car)
}

func (u *carUsecase) DeleteCar(ctx context.Context, id string) error {
	return u.carRepo.Delete(ctx, id)
}

func (u *carUsecase) ListCars(ctx context.Context) ([]*domain.Car, error) {
	return u.carRepo.List(ctx)
}

func (u *carUsecase) SearchCars(ctx context.Context, category string, minPrice, maxPrice float64, location string) ([]*domain.Car, error) {
	if minPrice > maxPrice && maxPrice != 0 {
		return nil, errors.New("min price cannot be greater than max price")
	}
	return u.carRepo.Search(ctx, category, minPrice, maxPrice, location)
}

func (u *carUsecase) CheckAvailability(ctx context.Context, id string, startDate, endDate string) (bool, error) {
	// In a real system, you would query a bookings table using the dates.
	// For this inventory scope, we verify the base availability flag.
	car, err := u.carRepo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return car.Available, nil
}

func (u *carUsecase) UpdateAvailability(ctx context.Context, id string, available bool) error {
	return u.carRepo.UpdateAvailability(ctx, id, available)
}

func (u *carUsecase) GetCarPricing(ctx context.Context, id string) (float64, error) {
	car, err := u.carRepo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return car.PricePerDay, nil
}

func (u *carUsecase) SetDynamicPricing(ctx context.Context, id string, newPrice float64) error {
	if newPrice <= 0 {
		return errors.New("new price must be strictly positive")
	}
	return u.carRepo.UpdatePricing(ctx, id, newPrice)
}

func (u *carUsecase) UploadCarImages(ctx context.Context, id string, urls []string) error {
	if len(urls) == 0 {
		return errors.New("no image URLs provided")
	}
	return u.carRepo.AddImages(ctx, id, urls)
}

func (u *carUsecase) GetCarReviews(ctx context.Context, id string) ([]*domain.Review, error) {
	return u.carRepo.GetReviews(ctx, id)
}
