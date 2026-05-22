package domain

import "context"

type EventPublisher interface {
	PublishCarCreated(ctx context.Context, car *Car) error
	PublishAvailabilityUpdated(ctx context.Context, id string, available bool) error
	PublishPricingUpdated(ctx context.Context, id string, price float64) error
}
