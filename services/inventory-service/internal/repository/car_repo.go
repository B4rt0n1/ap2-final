package postgres

import (
	"context"
	"database/sql"
	"inventory-service/internal/domain"

	"github.com/lib/pq"
)

type postgreCarRepository struct {
	db *sql.DB
}

func NewPostgreCarRepository(db *sql.DB) domain.CarRepository {
	return &postgreCarRepository{db: db}
}

func (r *postgreCarRepository) Create(ctx context.Context, car *domain.Car) error {
	query := `INSERT INTO cars (id, brand, model, year, category, price_per_day, available) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, car.ID, car.Brand, car.Model, car.Year, car.Category, car.PricePerDay, car.Available)
	return err
}

func (r *postgreCarRepository) GetByID(ctx context.Context, id string) (*domain.Car, error) {
	query := `SELECT id, brand, model, year, category, price_per_day, available FROM cars WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var c domain.Car
	err := row.Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.Category, &c.PricePerDay, &c.Available)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *postgreCarRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM cars WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgreCarRepository) Update(ctx context.Context, car *domain.Car) error {
	query := `UPDATE cars SET brand=$1, model=$2, price_per_day=$3, available=$4 WHERE id=$5`
	_, err := r.db.ExecContext(ctx, query, car.Brand, car.Model, car.PricePerDay, car.Available, car.ID)
	return err
}

func (r *postgreCarRepository) List(ctx context.Context) ([]*domain.Car, error) {
	query := `SELECT id, brand, model, year, category, price_per_day, available FROM cars`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []*domain.Car
	for rows.Next() {
		var c domain.Car
		if err := rows.Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.Category, &c.PricePerDay, &c.Available); err != nil {
			return nil, err
		}
		cars = append(cars, &c)
	}
	return cars, nil
}

func (r *postgreCarRepository) Search(ctx context.Context, category string, minPrice, maxPrice float64, location string) ([]*domain.Car, error) {
	query := `SELECT id, brand, model, year, category, price_per_day, available FROM cars 
	          WHERE category = $1 AND price_per_day >= $2 AND price_per_day <= $3`
	rows, err := r.db.QueryContext(ctx, query, category, minPrice, maxPrice)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []*domain.Car
	for rows.Next() {
		var c domain.Car
		if err := rows.Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.Category, &c.PricePerDay, &c.Available); err != nil {
			return nil, err
		}
		cars = append(cars, &c)
	}
	return cars, nil
}

func (r *postgreCarRepository) UpdateAvailability(ctx context.Context, id string, available bool) error {
	query := `UPDATE cars SET available = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, available, id)
	return err
}

func (r *postgreCarRepository) UpdatePricing(ctx context.Context, id string, newPrice float64) error {
	query := `UPDATE cars SET price_per_day = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, newPrice, id)
	return err
}

func (r *postgreCarRepository) AddImages(ctx context.Context, id string, urls []string) error {
	// Assuming cars table has an image_urls column of type TEXT[]
	query := `UPDATE cars SET image_urls = array_cat(image_urls, $1) WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, pq.Array(urls), id)
	return err
}

func (r *postgreCarRepository) GetReviews(ctx context.Context, id string) ([]*domain.Review, error) {
	// Assuming a separate reviews table linked by car_id
	query := `SELECT user_id, rating, comment FROM reviews WHERE car_id = $1`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		var rev domain.Review
		if err := rows.Scan(&rev.UserID, &rev.Rating, &rev.Comment); err != nil {
			return nil, err
		}
		reviews = append(reviews, &rev)
	}
	return reviews, nil
}
