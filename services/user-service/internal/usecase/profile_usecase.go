package usecase

import (
	"context"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
)

func (u *userUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *userUseCase) UpdateProfile(ctx context.Context, id, firstName, lastName, phone string) (*domain.User, error) {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone

	if err := u.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUseCase) UploadDriverLicense(ctx context.Context, id, licenseNumber, imageURL string) error {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.DriverLicenseNumber = licenseNumber
	user.LicenseImageURL = imageURL

	return u.repo.Update(ctx, user)
}

func (u *userUseCase) DeleteUser(ctx context.Context, id string) error {
	if _, err := u.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return u.repo.Delete(ctx, id)
}

func (u *userUseCase) GetRentalHistory(ctx context.Context, id string) ([]domain.RentalHistory, error) {
	if _, err := u.repo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	return []domain.RentalHistory{}, nil
}

func (u *userUseCase) GetPaymentMethods(ctx context.Context, id string) ([]domain.PaymentMethod, error) {
	if _, err := u.repo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	return []domain.PaymentMethod{}, nil
}
