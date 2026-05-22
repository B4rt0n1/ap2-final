package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginReturnsTokensForValidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	repo := &fakeUserRepository{
		byEmail: map[string]*domain.User{
			"ada@example.com": {
				ID:           "user-1",
				Email:        "ada@example.com",
				PasswordHash: string(hash),
			},
		},
	}
	uc := NewUserUseCase(repo, nil)

	token, refresh, err := uc.Login(context.Background(), "ada@example.com", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token != "dummy_access_token_user-1" {
		t.Fatalf("access token = %q, want dummy_access_token_user-1", token)
	}
	if refresh == "" {
		t.Fatal("refresh token is empty")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	repo := &fakeUserRepository{
		byEmail: map[string]*domain.User{
			"ada@example.com": {
				ID:           "user-1",
				Email:        "ada@example.com",
				PasswordHash: string(hash),
			},
		},
	}
	uc := NewUserUseCase(repo, nil)

	_, _, err = uc.Login(context.Background(), "ada@example.com", "wrong")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, domain.ErrInvalidCredentials)
	}
}

func TestRegisterRejectsExistingEmail(t *testing.T) {
	repo := &fakeUserRepository{
		byEmail: map[string]*domain.User{
			"ada@example.com": {ID: "user-1", Email: "ada@example.com"},
		},
	}
	uc := NewUserUseCase(repo, nil)

	_, err := uc.Register(context.Background(), "Ada", "Lovelace", "ada@example.com", "secret", "+100")
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("Register() error = %v, want %v", err, domain.ErrUserAlreadyExists)
	}
	if repo.created != nil {
		t.Fatalf("Register() created user despite duplicate: %#v", repo.created)
	}
}

func TestUpdateProfilePersistsChangedFields(t *testing.T) {
	repo := &fakeUserRepository{
		byID: map[string]*domain.User{
			"user-1": {
				ID:        "user-1",
				FirstName: "Old",
				LastName:  "Name",
				Phone:     "+100",
			},
		},
	}
	uc := NewUserUseCase(repo, nil)

	user, err := uc.UpdateProfile(context.Background(), "user-1", "Ada", "Lovelace", "+200")
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if user.FirstName != "Ada" || user.LastName != "Lovelace" || user.Phone != "+200" {
		t.Fatalf("UpdateProfile() user = %#v", user)
	}
	if repo.updated == nil || repo.updated.ID != "user-1" {
		t.Fatalf("UpdateProfile() did not persist update: %#v", repo.updated)
	}
}

func TestUploadDriverLicensePersistsLicenseFields(t *testing.T) {
	repo := &fakeUserRepository{
		byID: map[string]*domain.User{
			"user-1": {ID: "user-1"},
		},
	}
	uc := NewUserUseCase(repo, nil)

	if err := uc.UploadDriverLicense(context.Background(), "user-1", "DL-123", "https://example.com/license.png"); err != nil {
		t.Fatalf("UploadDriverLicense() error = %v", err)
	}
	if repo.updated.DriverLicenseNumber != "DL-123" || repo.updated.LicenseImageURL != "https://example.com/license.png" {
		t.Fatalf("updated license fields = %#v", repo.updated)
	}
}

type fakeUserRepository struct {
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
	created *domain.User
	updated *domain.User
}

func (r *fakeUserRepository) Create(_ context.Context, user *domain.User) error {
	r.created = cloneUser(user)
	return nil
}

func (r *fakeUserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (r *fakeUserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (r *fakeUserRepository) Update(_ context.Context, user *domain.User) error {
	r.updated = cloneUser(user)
	return nil
}

func (r *fakeUserRepository) Delete(context.Context, string) error {
	return nil
}

func cloneUser(user *domain.User) *domain.User {
	if user == nil {
		return nil
	}
	copied := *user
	return &copied
}
