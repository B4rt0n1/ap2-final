package usecase

import (
	"context"
	"testing"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterLoginVerifyAndChangePassword(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryUsers()
	service := NewUserUseCase(repo, nil)

	user, err := service.Register(ctx, "Aron", "Demo", "aron@example.com", "secret123", "+7700")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.ID == "" || user.VerificationCode == "" {
		t.Fatalf("Register() user = %#v", user)
	}

	if _, _, err := service.Login(ctx, user.Email, "secret123"); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if err := service.VerifyEmail(ctx, user.ID, "wrong"); err != domain.ErrInvalidVerificationCode {
		t.Fatalf("VerifyEmail() wrong code error = %v", err)
	}
	if err := service.VerifyEmail(ctx, user.ID, user.VerificationCode); err != nil {
		t.Fatalf("VerifyEmail() error = %v", err)
	}
	if err := service.ChangePassword(ctx, user.ID, "secret123", "new-secret"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if _, _, err := service.Login(ctx, user.Email, "new-secret"); err != nil {
		t.Fatalf("Login() changed password error = %v", err)
	}
}

func TestProfileLicenseAndReferenceEndpoints(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryUsers()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}
	repo.users["user-1"] = &domain.User{
		ID:           "user-1",
		FirstName:    "Before",
		LastName:     "Name",
		Email:        "profile@example.com",
		PasswordHash: string(hash),
	}

	service := NewUserUseCase(repo, nil)
	updated, err := service.UpdateProfile(ctx, "user-1", "After", "Demo", "+7711")
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if updated.FirstName != "After" || updated.Phone != "+7711" {
		t.Fatalf("UpdateProfile() user = %#v", updated)
	}
	if err := service.UploadDriverLicense(ctx, "user-1", "DL-42", "https://example.com/license.png"); err != nil {
		t.Fatalf("UploadDriverLicense() error = %v", err)
	}
	if rentals, err := service.GetRentalHistory(ctx, "user-1"); err != nil || len(rentals) != 0 {
		t.Fatalf("GetRentalHistory() rentals = %#v, err = %v", rentals, err)
	}
	if methods, err := service.GetPaymentMethods(ctx, "user-1"); err != nil || len(methods) != 0 {
		t.Fatalf("GetPaymentMethods() methods = %#v, err = %v", methods, err)
	}
	if err := service.Logout(ctx, "user-1"); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if err := service.DeleteUser(ctx, "user-1"); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if _, err := service.GetUser(ctx, "user-1"); err != domain.ErrUserNotFound {
		t.Fatalf("GetUser() after delete error = %v", err)
	}
}

type memoryUsers struct {
	users map[string]*domain.User
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{users: make(map[string]*domain.User)}
}

func (m *memoryUsers) Create(_ context.Context, user *domain.User) error {
	m.users[user.ID] = cloneMemoryUser(user)
	return nil
}

func (m *memoryUsers) GetByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return cloneMemoryUser(user), nil
}

func (m *memoryUsers) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return cloneMemoryUser(user), nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *memoryUsers) Update(_ context.Context, user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	m.users[user.ID] = cloneMemoryUser(user)
	return nil
}

func (m *memoryUsers) Delete(_ context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return domain.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func cloneMemoryUser(user *domain.User) *domain.User {
	copied := *user
	return &copied
}
