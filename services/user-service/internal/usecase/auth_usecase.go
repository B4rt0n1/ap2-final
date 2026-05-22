package usecase

import (
	"context"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"golang.org/x/crypto/bcrypt"
)

type userUseCase struct {
	repo     domain.UserRepository
	natsConn *nats.Conn
}

func NewUserUseCase(repo domain.UserRepository, nc *nats.Conn) domain.UserUseCase {
	return &userUseCase{
		repo:     repo,
		natsConn: nc,
	}
}

func (u *userUseCase) Register(ctx context.Context, firstName, lastName, email, password, phone string) (*domain.User, error) {
	if password == "" {
		return nil, domain.ErrInvalidPassword
	}

	existing, _ := u.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		ID:               uuid.NewString(),
		FirstName:        firstName,
		LastName:         lastName,
		Email:            email,
		PasswordHash:     string(hash),
		Phone:            phone,
		VerificationCode: uuid.NewString(),
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	if u.natsConn != nil {
		_ = u.natsConn.Publish("user.registered", []byte(user.Email))
	}

	return user, nil
}

func (u *userUseCase) Logout(ctx context.Context, id string) error {
	_, err := u.repo.GetByID(ctx, id)
	return err
}

func (u *userUseCase) VerifyEmail(ctx context.Context, id, code string) error {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if code == "" || user.VerificationCode != code {
		return domain.ErrInvalidVerificationCode
	}

	user.IsEmailVerified = true
	user.VerificationCode = ""
	return u.repo.Update(ctx, user)
}

func (u *userUseCase) ResetPassword(ctx context.Context, email string) error {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	user.VerificationCode = uuid.NewString()
	if err := u.repo.Update(ctx, user); err != nil {
		return err
	}
	if u.natsConn != nil {
		_ = u.natsConn.Publish("user.password_reset_requested", []byte(user.Email))
	}
	return nil
}

func (u *userUseCase) ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	if newPassword == "" {
		return domain.ErrInvalidPassword
	}

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternal
	}
	user.PasswordHash = string(hash)
	return u.repo.Update(ctx, user)
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	token := "dummy_access_token_" + user.ID
	refreshToken := "dummy_refresh_token"

	return token, refreshToken, nil
}
