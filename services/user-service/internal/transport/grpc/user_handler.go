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
