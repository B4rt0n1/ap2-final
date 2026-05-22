package domain

import (
	"context"
	"time"
)

type User struct {
	ID                  string
	FirstName           string
	LastName            string
	Email               string
	PasswordHash        string
	Phone               string
	DriverLicenseNumber string
	LicenseImageURL     string
	IsEmailVerified     bool
	VerificationCode    string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

type UserUseCase interface {
	Register(ctx context.Context, firstName, lastName, email, password, phone string) (*User, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	GetUser(ctx context.Context, id string) (*User, error)
	UpdateProfile(ctx context.Context, id, firstName, lastName, phone string) (*User, error)
	UploadDriverLicense(ctx context.Context, id, licenseNumber, imageURL string) error
}
