package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
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

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	user := &domain.User{
		ID:           "integration-user-" + suffix,
		FirstName:    "Ada",
		LastName:     "Lovelace",
		Email:        "ada-" + suffix + "@example.com",
		PasswordHash: "hash",
		Phone:        "+100",
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID); err != nil {
			t.Errorf("cleanup user: %v", err)
		}
	})

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	byEmail, err := repo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if byEmail.ID != user.ID || byEmail.Phone != "+100" {
		t.Fatalf("GetByEmail() user = %#v", byEmail)
	}

	byEmail.FirstName = "Augusta"
	byEmail.DriverLicenseNumber = "DL-123"
	byEmail.LicenseImageURL = "https://example.com/license.png"
	if err := repo.Update(ctx, byEmail); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if updated.FirstName != "Augusta" || updated.DriverLicenseNumber != "DL-123" {
		t.Fatalf("updated user = %#v", updated)
	}

	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(ctx, user.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func openUserIntegrationDatabase(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() error = %v", err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext() error = %v", err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "000001_create_users_table.up.sql"))
	if err != nil {
		t.Fatalf("read user migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil && !isUserRelationAlreadyExists(err) {
		t.Fatalf("apply user migration: %v", err)
	}

	return db
}

func isUserRelationAlreadyExists(err error) bool {
	return err != nil && userSQLState(err) == "42P07"
}

type userSQLStateError interface {
	SQLState() string
}

func userSQLState(err error) string {
	var value userSQLStateError
	if errors.As(err, &value) {
		return value.SQLState()
	}

	return ""
}
