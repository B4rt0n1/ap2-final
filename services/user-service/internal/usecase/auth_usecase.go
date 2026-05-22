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
	existing, _ := u.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: string(hash),
		Phone:        phone,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	u.natsConn.Publish("user.registered", []byte(user.Email))

	return user, nil
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
