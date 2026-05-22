package domain

import "context"

type Car struct {
	ID          string
	Brand       string
	Model       string
	Year        int32
	Category    string
	PricePerDay float64
	Available   bool
	ImageURLs   []string // Added for UploadCarImages
}

type Review struct {
	UserID  string
	Rating  int32
	Comment string
}

type CarRepository interface {
	Create(ctx context.Context, car *Car) error
	GetByID(ctx context.Context, id string) (*Car, error)
	Update(ctx context.Context, car *Car) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*Car, error)
	Search(ctx context.Context, category string, minPrice, maxPrice float64, location string) ([]*Car, error)
	UpdateAvailability(ctx context.Context, id string, available bool) error
	UpdatePricing(ctx context.Context, id string, newPrice float64) error
	AddImages(ctx context.Context, id string, urls []string) error
	GetReviews(ctx context.Context, id string) ([]*Review, error)
}

type CarUsecase interface {
	AddCar(ctx context.Context, car *Car) error
	GetCarByID(ctx context.Context, id string) (*Car, error)
	UpdateCar(ctx context.Context, car *Car) error
	DeleteCar(ctx context.Context, id string) error
	ListCars(ctx context.Context) ([]*Car, error)
	SearchCars(ctx context.Context, category string, minPrice, maxPrice float64, location string) ([]*Car, error)
	CheckAvailability(ctx context.Context, id string, startDate, endDate string) (bool, error)
	UpdateAvailability(ctx context.Context, id string, available bool) error
	GetCarPricing(ctx context.Context, id string) (float64, error)
	SetDynamicPricing(ctx context.Context, id string, newPrice float64) error
	UploadCarImages(ctx context.Context, id string, urls []string) error
	GetCarReviews(ctx context.Context, id string) ([]*Review, error)
}
