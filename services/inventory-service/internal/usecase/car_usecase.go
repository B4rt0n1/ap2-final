package usecase

import (
	"context"
	"errors"
	"inventory-service/internal/domain"
	"log"
)

type carUsecase struct {
	carRepo   domain.CarRepository
	carCache  domain.CarCache
	publisher domain.EventPublisher
}

func NewCarUsecase(repo domain.CarRepository, cache domain.CarCache, pub domain.EventPublisher) domain.CarUsecase {
	return &carUsecase{
		carRepo:   repo,
		carCache:  cache,
		publisher: pub,
	}
}

func (u *carUsecase) AddCar(ctx context.Context, car *domain.Car) error {
	if car.PricePerDay <= 0 {
		return errors.New("price per day must be strictly positive")
	}

	if err := u.carRepo.Create(ctx, car); err != nil {
		return err
	}

	// Asynchronously publish to NATS queue
	go func() {
		if err := u.publisher.PublishCarCreated(context.Background(), car); err != nil {
			log.Printf("failed to publish car created event: %v", err)
		}
	}()

	return nil
}

func (u *carUsecase) GetCarByID(ctx context.Context, id string) (*domain.Car, error) {
	cachedCar, err := u.carCache.Get(ctx, id)
	if err == nil && cachedCar != nil {
		return cachedCar, nil
	}

	// Cache miss: Query Postgres database
	car, err := u.carRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Populate cache for subsequent lookups
	go func() {
		_ = u.carCache.Set(context.Background(), car)
	}()

	return car, nil
}

func (u *carUsecase) UpdateCar(ctx context.Context, car *domain.Car) error {
	if err := u.carRepo.Update(ctx, car); err != nil {
		return err
	}
	// Invalidate cache on update
	_ = u.carCache.Delete(ctx, car.ID)
	return nil
}

func (u *carUsecase) DeleteCar(ctx context.Context, id string) error {
	if err := u.carRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = u.carCache.Delete(ctx, id)
	return nil
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
	if err := u.carRepo.UpdateAvailability(ctx, id, available); err != nil {
		return err
	}

	// Evict outdated cache entry
	_ = u.carCache.Delete(ctx, id)

	// Stream update to consumer microservices (e.g., Booking Service) via NATS
	go func() {
		_ = u.publisher.PublishAvailabilityUpdated(context.Background(), id, available)
	}()

	return nil
}

func (u *carUsecase) GetCarPricing(ctx context.Context, id string) (float64, error) {
	car, err := u.carRepo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return car.PricePerDay, nil
}

func (u *carUsecase) SetDynamicPricing(ctx context.Context, id string, newPrice float64) error {
	if err := u.carRepo.UpdatePricing(ctx, id, newPrice); err != nil {
		return err
	}

	_ = u.carCache.Delete(ctx, id)

	go func() {
		_ = u.publisher.PublishPricingUpdated(context.Background(), id, newPrice)
	}()

	return nil
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
