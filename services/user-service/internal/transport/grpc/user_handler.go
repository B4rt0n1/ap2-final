package grpcserver

import (
	"context"
	"errors"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	useCase domain.UserUseCase
}

func NewUserHandler(uc domain.UserUseCase) *UserHandler {
	return &UserHandler{useCase: uc}
}

func mapToProtoUser(user *domain.User) *userv1.User {
	return &userv1.User{
		Id:                  user.ID,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		Phone:               user.Phone,
		DriverLicenseNumber: user.DriverLicenseNumber,
	}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.UserResponse, error) {
	user, err := h.useCase.Register(ctx, req.GetFirstName(), req.GetLastName(), req.GetEmail(), req.GetPassword(), req.GetPhone())
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.UserResponse{User: mapToProtoUser(user)}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *userv1.LoginUserRequest) (*userv1.AuthResponse, error) {
	token, refresh, err := h.useCase.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.AuthResponse{
		Token:        token,
		RefreshToken: refresh,
	}, nil
}

func (h *UserHandler) LogoutUser(ctx context.Context, req *userv1.LogoutUserRequest) (*userv1.MessageResponse, error) {
	if err := h.useCase.Logout(ctx, req.GetUserId()); err != nil {
		return nil, userError(err)
	}
	return &userv1.MessageResponse{Message: "user logged out"}, nil
}

func (h *UserHandler) GetUserById(ctx context.Context, req *userv1.GetUserByIdRequest) (*userv1.UserResponse, error) {
	user, err := h.useCase.GetUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.UserResponse{User: mapToProtoUser(user)}, nil
}

func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *userv1.UpdateUserProfileRequest) (*userv1.UserResponse, error) {
	user, err := h.useCase.UpdateProfile(ctx, req.GetUserId(), req.GetFirstName(), req.GetLastName(), req.GetPhone())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.UserResponse{User: mapToProtoUser(user)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.MessageResponse, error) {
	if err := h.useCase.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, userError(err)
	}
	return &userv1.MessageResponse{Message: "user deleted"}, nil
}

func (h *UserHandler) VerifyEmail(ctx context.Context, req *userv1.VerifyEmailRequest) (*userv1.MessageResponse, error) {
	if err := h.useCase.VerifyEmail(ctx, req.GetUserId(), req.GetVerificationCode()); err != nil {
		return nil, userError(err)
	}
	return &userv1.MessageResponse{Message: "email verified"}, nil
}

func (h *UserHandler) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*userv1.MessageResponse, error) {
	if err := h.useCase.ResetPassword(ctx, req.GetEmail()); err != nil {
		return nil, userError(err)
	}
	return &userv1.MessageResponse{Message: "password reset requested"}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*userv1.MessageResponse, error) {
	if err := h.useCase.ChangePassword(ctx, req.GetUserId(), req.GetOldPassword(), req.GetNewPassword()); err != nil {
		return nil, userError(err)
	}
	return &userv1.MessageResponse{Message: "password changed"}, nil
}

func (h *UserHandler) UploadDriverLicense(ctx context.Context, req *userv1.UploadDriverLicenseRequest) (*userv1.MessageResponse, error) {
	err := h.useCase.UploadDriverLicense(ctx, req.GetUserId(), req.GetLicenseNumber(), req.GetLicenseImageUrl())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.MessageResponse{Message: "Driver license uploaded successfully"}, nil
}

func (h *UserHandler) GetUserRentalHistory(ctx context.Context, req *userv1.GetUserRentalHistoryRequest) (*userv1.RentalHistoryResponse, error) {
	rentals, err := h.useCase.GetRentalHistory(ctx, req.GetUserId())
	if err != nil {
		return nil, userError(err)
	}

	response := &userv1.RentalHistoryResponse{Rentals: make([]*userv1.RentalHistory, 0, len(rentals))}
	for _, rental := range rentals {
		response.Rentals = append(response.Rentals, &userv1.RentalHistory{
			BookingId: rental.BookingID,
			CarId:     rental.CarID,
			StartDate: rental.StartDate,
			EndDate:   rental.EndDate,
		})
	}
	return response, nil
}

func (h *UserHandler) GetUserPaymentMethods(ctx context.Context, req *userv1.GetUserPaymentMethodsRequest) (*userv1.PaymentMethodsResponse, error) {
	methods, err := h.useCase.GetPaymentMethods(ctx, req.GetUserId())
	if err != nil {
		return nil, userError(err)
	}

	response := &userv1.PaymentMethodsResponse{Methods: make([]*userv1.PaymentMethod, 0, len(methods))}
	for _, method := range methods {
		response.Methods = append(response.Methods, &userv1.PaymentMethod{
			Id:             method.ID,
			Type:           method.Type,
			LastFourDigits: method.LastFourDigits,
		})
	}
	return response, nil
}

func userError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrInvalidVerificationCode),
		errors.Is(err, domain.ErrInvalidPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
