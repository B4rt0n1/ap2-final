package usecase

import "errors"

var (
	ErrIDGeneratorUnavailable = errors.New("booking id generator is unavailable")
	ErrPricingUnavailable     = errors.New("booking pricing provider is unavailable")
	ErrTransactionUnavailable = errors.New("booking transaction manager is unavailable")
	ErrInvalidRentalDays      = errors.New("rental days must be greater than zero")
	ErrInvalidDailyPrice      = errors.New("daily price cannot be negative")
	ErrInvalidDiscountCode    = errors.New("discount code is invalid")
	ErrDiscountNotApplicable  = errors.New("discount can only be applied to pending bookings")
	ErrBookingNotEditable     = errors.New("booking dates can only be updated while pending")
	ErrInvalidIssue           = errors.New("booking issue description is required")
	ErrUserValidationMissing  = errors.New("user validator is unavailable")
	ErrCarValidationMissing   = errors.New("car availability checker is unavailable")
	ErrCarUnavailable         = errors.New("car is not available for requested rental dates")
)
