package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"inventory-service/internal/domain"

	_ "github.com/lib/pq"
)

func TestCarRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("INVENTORY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("INVENTORY_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db := openCarIntegrationDatabase(t, ctx, dsn)
	repo := NewPostgreCarRepository(db)
	carID := "integration-car-" + time.Now().UTC().Format("20060102150405.000000000")
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM cars WHERE id = $1", carID); err != nil {
			t.Errorf("cleanup car: %v", err)
		}
	})

	car := &domain.Car{
		ID:          carID,
		Brand:       "Toyota",
		Model:       "Camry",
		Year:        2025,
		Category:    "sedan",
		PricePerDay: 85,
		Available:   true,
	}
	if err := repo.Create(ctx, car); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := repo.GetByID(ctx, carID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Brand != car.Brand || got.PricePerDay != car.PricePerDay {
		t.Fatalf("GetByID() car = %#v", got)
	}

	if err := repo.UpdatePricing(ctx, carID, 99); err != nil {
		t.Fatalf("UpdatePricing() error = %v", err)
	}
	priced, err := repo.GetByID(ctx, carID)
	if err != nil || priced.PricePerDay != 99 {
		t.Fatalf("updated pricing car = %#v, err = %v", priced, err)
	}
}

func openCarIntegrationDatabase(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext() error = %v", err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "000001_create_inventory_table.up.sql"))
	if err != nil {
		t.Fatalf("read inventory migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("apply inventory migration: %v", err)
	}
	return db
}
