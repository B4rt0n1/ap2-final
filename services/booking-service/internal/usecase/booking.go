package usecase

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/domain"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/repository"
)

// IDGenerator provides booking identifiers before persistence.
type IDGenerator interface {
	NewID() string
}

// PricingProvider gives booking flows the inventory daily rental rate.
type PricingProvider interface {
	DailyPrice(ctx context.Context, carID string) (float64, error)
}

// UserValidator confirms that a booking user exists in the user service boundary.
type UserValidator interface {
	EnsureUserExists(ctx context.Context, userID string) error
}

// CarAvailabilityChecker confirms that inventory can rent a car for a window.
type CarAvailabilityChecker interface {
	EnsureCarAvailable(ctx context.Context, carID string, startDate, endDate time.Time) error
}

// IssueReporter sends customer-reported booking issues to the next boundary.
type IssueReporter interface {
	ReportBookingIssue(ctx context.Context, issue IssueReport) error
}

// Dependencies groups the booking use case boundaries.
type Dependencies struct {
	Bookings    repository.BookingStore
	Transactor  repository.Transactor
	IDs         IDGenerator
	Pricing     PricingProvider
	Users       UserValidator
	Cars        CarAvailabilityChecker
	Events      EventPublisher
	IssueReport IssueReporter
}

// Service coordinates booking application flows.
type Service struct {
	bookings    repository.BookingStore
	transactor  repository.Transactor
	ids         IDGenerator
	pricing     PricingProvider
	users       UserValidator
	cars        CarAvailabilityChecker
	events      EventPublisher
	issueReport IssueReporter
}

// CreateBookingInput contains the data needed to open a rental reservation.
type CreateBookingInput struct {
	UserID    string
	CarID     string
	StartDate time.Time
	EndDate   time.Time
}

// UpdateBookingInput contains the dates a pending booking can move to.
type UpdateBookingInput struct {
	BookingID string
	StartDate time.Time
	EndDate   time.Time
}

// RentalCost is the current cost calculation result.
type RentalCost struct {
	OriginalPrice   float64
	DiscountedPrice float64
}

// IssueReport carries a customer problem tied to a booking.
type IssueReport struct {
	BookingID   string
	Description string
	ReportedAt  time.Time
}

// New creates a booking use case service.
func New(deps Dependencies) *Service {
	return &Service{
		bookings:    deps.Bookings,
		transactor:  deps.Transactor,
		ids:         deps.IDs,
		pricing:     deps.Pricing,
		users:       deps.Users,
		cars:        deps.Cars,
		events:      deps.Events,
		issueReport: deps.IssueReport,
	}
}

// CreateBooking creates a pending booking with its inventory-derived cost.
func (s *Service) CreateBooking(ctx context.Context, input CreateBookingInput) (domain.Booking, error) {
	booking, err := domain.NewBooking(input.UserID, input.CarID, input.StartDate, input.EndDate, 0)
	if err != nil {
		return domain.Booking{}, err
	}

	if err := s.ensureUserExists(ctx, input.UserID); err != nil {
		return domain.Booking{}, err
	}
	if err := s.ensureCarAvailable(ctx, input.CarID, input.StartDate, input.EndDate); err != nil {
		return domain.Booking{}, err
	}

	booking.ID, err = s.newBookingID()
	if err != nil {
		return domain.Booking{}, err
	}

	cost, err := s.CalculateRentalCost(ctx, input.CarID, rentalDays(input.StartDate, input.EndDate))
	if err != nil {
		return domain.Booking{}, err
	}
	booking.TotalPrice = cost.DiscountedPrice

	var created domain.Booking
	if err := s.withinTransaction(ctx, func(ctx context.Context, store repository.BookingStore) error {
		var createErr error
		created, createErr = store.Create(ctx, booking)
		return createErr
	}); err != nil {
		return domain.Booking{}, err
	}

	if err := s.publishEvent(ctx, EventBookingCreated, created); err != nil {
		return domain.Booking{}, err
	}

	return created, nil
}

// GetBookingByID returns a booking by identifier.
func (s *Service) GetBookingByID(ctx context.Context, bookingID string) (domain.Booking, error) {
	if err := validateBookingID(bookingID); err != nil {
		return domain.Booking{}, err
	}

	return s.bookings.GetByID(ctx, bookingID)
}

// ListUserBookings returns the user's bookings.
func (s *Service) ListUserBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrInvalidUserID
	}

	return s.bookings.ListByUserID(ctx, userID)
}

// UpdateBooking changes dates while a booking is still pending.
func (s *Service) UpdateBooking(ctx context.Context, input UpdateBookingInput) (domain.Booking, error) {
	if err := validateBookingID(input.BookingID); err != nil {
		return domain.Booking{}, err
	}

	var updated domain.Booking
	if err := s.withinTransaction(ctx, func(ctx context.Context, store repository.BookingStore) error {
		booking, err := store.GetByID(ctx, input.BookingID)
		if err != nil {
			return err
		}
		if booking.Status != domain.StatusPending {
			return ErrBookingNotEditable
		}

		booking.StartDate = input.StartDate
		booking.EndDate = input.EndDate
		if err := booking.Validate(); err != nil {
			return err
		}

		updated, err = store.Update(ctx, booking)
		return err
	}); err != nil {
		return domain.Booking{}, err
	}

	return updated, nil
}

// ConfirmBooking moves a pending booking into its confirmed state.
func (s *Service) ConfirmBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return s.transitionBooking(ctx, bookingID, domain.StatusConfirmed, EventBookingConfirmed)
}

// CancelBooking cancels a pending or confirmed booking.
func (s *Service) CancelBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return s.transitionBooking(ctx, bookingID, domain.StatusCancelled, EventBookingCancelled)
}

// StartRental marks a confirmed booking as an active rental.
func (s *Service) StartRental(ctx context.Context, bookingID string) (domain.Booking, error) {
	return s.transitionBooking(ctx, bookingID, domain.StatusActive, "")
}

// EndRental completes an active rental.
func (s *Service) EndRental(ctx context.Context, bookingID string) (domain.Booking, error) {
	return s.transitionBooking(ctx, bookingID, domain.StatusCompleted, "")
}

// CalculateRentalCost prices a car for a count of rental days.
func (s *Service) CalculateRentalCost(ctx context.Context, carID string, days int) (RentalCost, error) {
	if strings.TrimSpace(carID) == "" {
		return RentalCost{}, domain.ErrInvalidCarID
	}
	if days <= 0 {
		return RentalCost{}, ErrInvalidRentalDays
	}
	if s.pricing == nil {
		return RentalCost{}, ErrPricingUnavailable
	}

	dailyPrice, err := s.pricing.DailyPrice(ctx, carID)
	if err != nil {
		return RentalCost{}, err
	}
	if dailyPrice < 0 {
		return RentalCost{}, ErrInvalidDailyPrice
	}

	total := roundMoney(dailyPrice * float64(days))
	return RentalCost{
		OriginalPrice:   total,
		DiscountedPrice: total,
	}, nil
}

// ApplyDiscount applies a supported promotion to a pending booking.
func (s *Service) ApplyDiscount(ctx context.Context, bookingID, code string) (RentalCost, error) {
	if err := validateBookingID(bookingID); err != nil {
		return RentalCost{}, err
	}

	rate, ok := discountRate(code)
	if !ok {
		return RentalCost{}, ErrInvalidDiscountCode
	}

	var cost RentalCost
	if err := s.withinTransaction(ctx, func(ctx context.Context, store repository.BookingStore) error {
		booking, err := store.GetByID(ctx, bookingID)
		if err != nil {
			return err
		}
		if booking.Status != domain.StatusPending {
			return ErrDiscountNotApplicable
		}

		cost.OriginalPrice = booking.TotalPrice
		cost.DiscountedPrice = roundMoney(booking.TotalPrice * (1 - rate))
		booking.TotalPrice = cost.DiscountedPrice

		_, err = store.Update(ctx, booking)
		return err
	}); err != nil {
		return RentalCost{}, err
	}

	return cost, nil
}

// GetBookingStatus returns the current lifecycle state.
func (s *Service) GetBookingStatus(ctx context.Context, bookingID string) (domain.Status, error) {
	booking, err := s.GetBookingByID(ctx, bookingID)
	if err != nil {
		return "", err
	}

	return booking.Status, nil
}

// ReportIssue validates a booking issue and forwards it to the configured reporter.
func (s *Service) ReportIssue(ctx context.Context, bookingID, description string) error {
	if err := validateBookingID(bookingID); err != nil {
		return err
	}
	if strings.TrimSpace(description) == "" {
		return ErrInvalidIssue
	}

	if _, err := s.bookings.GetByID(ctx, bookingID); err != nil {
		return err
	}
	if s.issueReport == nil {
		return nil
	}

	return s.issueReport.ReportBookingIssue(ctx, IssueReport{
		BookingID:   bookingID,
		Description: strings.TrimSpace(description),
		ReportedAt:  time.Now().UTC(),
	})
}

func (s *Service) transitionBooking(
	ctx context.Context,
	bookingID string,
	next domain.Status,
	eventType EventType,
) (domain.Booking, error) {
	if err := validateBookingID(bookingID); err != nil {
		return domain.Booking{}, err
	}

	var updated domain.Booking
	if err := s.withinTransaction(ctx, func(ctx context.Context, store repository.BookingStore) error {
		booking, err := store.GetByID(ctx, bookingID)
		if err != nil {
			return err
		}
		if err := booking.TransitionTo(next); err != nil {
			return err
		}

		updated, err = store.Update(ctx, booking)
		return err
	}); err != nil {
		return domain.Booking{}, err
	}

	if eventType != "" {
		if err := s.publishEvent(ctx, eventType, updated); err != nil {
			return domain.Booking{}, err
		}
	}

	return updated, nil
}

func (s *Service) withinTransaction(
	ctx context.Context,
	fn func(context.Context, repository.BookingStore) error,
) error {
	if s.transactor == nil {
		return ErrTransactionUnavailable
	}

	return s.transactor.WithinTransaction(ctx, fn)
}

func (s *Service) newBookingID() (string, error) {
	if s.ids == nil {
		return "", ErrIDGeneratorUnavailable
	}

	id := strings.TrimSpace(s.ids.NewID())
	if id == "" {
		return "", ErrIDGeneratorUnavailable
	}

	return id, nil
}

func (s *Service) ensureUserExists(ctx context.Context, userID string) error {
	if s.users == nil {
		return ErrUserValidationMissing
	}

	return s.users.EnsureUserExists(ctx, userID)
}

func (s *Service) ensureCarAvailable(ctx context.Context, carID string, startDate, endDate time.Time) error {
	if s.cars == nil {
		return ErrCarValidationMissing
	}

	return s.cars.EnsureCarAvailable(ctx, carID, startDate, endDate)
}

func (s *Service) publishEvent(ctx context.Context, eventType EventType, booking domain.Booking) error {
	if s.events == nil {
		return ErrEventPublisherMissing
	}

	return s.events.PublishBookingEvent(ctx, newBookingEvent(eventType, booking))
}

func validateBookingID(bookingID string) error {
	if strings.TrimSpace(bookingID) == "" {
		return domain.ErrInvalidBookingID
	}

	return nil
}

func rentalDays(startDate, endDate time.Time) int {
	return int(math.Ceil(endDate.Sub(startDate).Hours() / 24))
}

func discountRate(code string) (float64, bool) {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "RENT10":
		return 0.10, true
	case "WEEK15":
		return 0.15, true
	default:
		return 0, false
	}
}

func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}
