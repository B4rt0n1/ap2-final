package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestUserRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("USER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("USER_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db := openUserIntegrationDatabase(t, ctx, dsn)
	repo := NewUserRepository(db)
	userID := "integration-user-" + time.Now().UTC().Format("20060102150405.000000000")
	email := userID + "@example.com"
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID); err != nil {
			t.Errorf("cleanup user: %v", err)
		}
	})

	created := &domain.User{
		ID:               userID,
		FirstName:        "Integration",
		LastName:         "User",
		Email:            email,
		PasswordHash:     "hash",
		Phone:            "+7700",
		VerificationCode: "verify-me",
	}
	if err := repo.Create(ctx, created); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	got.IsEmailVerified = true
	got.VerificationCode = ""
	got.DriverLicenseNumber = "DL-42"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !updated.IsEmailVerified || updated.DriverLicenseNumber != "DL-42" {
		t.Fatalf("updated user = %#v", updated)
	}
}

func openUserIntegrationDatabase(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext() error = %v", err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "000001_create_users_table.up.sql"))
	if err != nil {
		t.Fatalf("read user migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("apply user migration: %v", err)
	}
	return db
}
