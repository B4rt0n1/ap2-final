package domain

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("user with this email already exists")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrInvalidVerificationCode = errors.New("invalid email verification code")
	ErrInvalidPassword         = errors.New("password must not be empty")
	ErrInternal                = errors.New("internal server error")
)
