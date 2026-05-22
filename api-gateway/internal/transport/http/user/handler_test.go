package userhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegisterCallsGRPCAndReturnsCreated(t *testing.T) {
	client := &recordingUserClient{
		registerResponse: &userv1.UserResponse{User: &userv1.User{Id: "user-1"}},
	}
	server := userMux(client)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","password":"secret","phone":"+100"}`))
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", response.Code, http.StatusCreated)
	}
	if client.registerRequest.GetEmail() != "ada@example.com" || client.registerRequest.GetFirstName() != "Ada" {
		t.Fatalf("register request = %#v", client.registerRequest)
	}
}

func TestRegisterRejectsInvalidJSON(t *testing.T) {
	server := userMux(&recordingUserClient{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{bad`))
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("register status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestLoginMapsGRPCErrorToUnauthorized(t *testing.T) {
	server := userMux(&recordingUserClient{
		loginErr: status.Error(codes.Unauthenticated, "invalid email or password"),
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"ada@example.com","password":"wrong"}`))
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("login status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func userMux(client userv1.UserServiceClient) http.Handler {
	mux := http.NewServeMux()
	Register(mux, client)
	return mux
}

type recordingUserClient struct {
	registerRequest  *userv1.RegisterUserRequest
	registerResponse *userv1.UserResponse
	loginRequest     *userv1.LoginUserRequest
	loginResponse    *userv1.AuthResponse
	loginErr         error
}

func (c *recordingUserClient) RegisterUser(_ context.Context, request *userv1.RegisterUserRequest, _ ...grpc.CallOption) (*userv1.UserResponse, error) {
	c.registerRequest = request
	return c.registerResponse, nil
}

func (c *recordingUserClient) LoginUser(_ context.Context, request *userv1.LoginUserRequest, _ ...grpc.CallOption) (*userv1.AuthResponse, error) {
	c.loginRequest = request
	if c.loginErr != nil {
		return nil, c.loginErr
	}
	if c.loginResponse != nil {
		return c.loginResponse, nil
	}
	return &userv1.AuthResponse{Token: "token", RefreshToken: "refresh"}, nil
}

func (c *recordingUserClient) LogoutUser(context.Context, *userv1.LogoutUserRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) GetUserById(context.Context, *userv1.GetUserByIdRequest, ...grpc.CallOption) (*userv1.UserResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) UpdateUserProfile(context.Context, *userv1.UpdateUserProfileRequest, ...grpc.CallOption) (*userv1.UserResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) DeleteUser(context.Context, *userv1.DeleteUserRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) VerifyEmail(context.Context, *userv1.VerifyEmailRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) ResetPassword(context.Context, *userv1.ResetPasswordRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) ChangePassword(context.Context, *userv1.ChangePasswordRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) UploadDriverLicense(context.Context, *userv1.UploadDriverLicenseRequest, ...grpc.CallOption) (*userv1.MessageResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) GetUserRentalHistory(context.Context, *userv1.GetUserRentalHistoryRequest, ...grpc.CallOption) (*userv1.RentalHistoryResponse, error) {
	return nil, nil
}

func (c *recordingUserClient) GetUserPaymentMethods(context.Context, *userv1.GetUserPaymentMethodsRequest, ...grpc.CallOption) (*userv1.PaymentMethodsResponse, error) {
	return nil, nil
}
