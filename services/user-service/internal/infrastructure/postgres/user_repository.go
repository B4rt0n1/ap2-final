package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, first_name, last_name, email, password_hash, phone, verification_code, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Phone, user.VerificationCode)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET first_name = $2,
		    last_name = $3,
		    phone = $4,
		    password_hash = $5,
		    driver_license_number = $6,
		    license_image_url = $7,
		    is_email_verified = $8,
		    verification_code = $9,
		    updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.PasswordHash,
		user.DriverLicenseNumber,
		user.LicenseImageURL,
		user.IsEmailVerified,
		user.VerificationCode,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, first_name, last_name, email, password_hash, phone, driver_license_number, license_image_url, is_email_verified, verification_code FROM users WHERE id = $1`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, first_name, last_name, email, password_hash, phone, driver_license_number, license_image_url, is_email_verified, verification_code FROM users WHERE email = $1`
	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var phone, driverLicense, licenseImage, verificationCode sql.NullString

	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &phone, &driverLicense, &licenseImage, &u.IsEmailVerified, &verificationCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	u.Phone = phone.String
	u.DriverLicenseNumber = driverLicense.String
	u.LicenseImageURL = licenseImage.String
	u.VerificationCode = verificationCode.String

	return &u, nil
}
