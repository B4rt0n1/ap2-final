package nats

import (
	"context"
	"encoding/json"

	"inventory-service/internal/domain"

	"github.com/nats-io/nats.go"
)

type natsEventPublisher struct {
	nc *nats.Conn
}

func NewNatsEventPublisher(nc *nats.Conn) domain.EventPublisher {
	return &natsEventPublisher{nc: nc}
}

type CarAvailabilityEvent struct {
	CarID     string `json:"car_id"`
	Available bool   `json:"available"`
}

type CarPriceEvent struct {
	CarID string  `json:"car_id"`
	Price float64 `json:"price"`
}

func (p *natsEventPublisher) PublishCarCreated(ctx context.Context, car *domain.Car) error {
	data, err := json.Marshal(car)
	if err != nil {
		return err
	}
	return p.nc.Publish("inventory.car.created", data)
}

func (p *natsEventPublisher) PublishAvailabilityUpdated(ctx context.Context, id string, available bool) error {
	event := CarAvailabilityEvent{CarID: id, Available: available}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.nc.Publish("inventory.car.availability_updated", data)
}

func (p *natsEventPublisher) PublishPricingUpdated(ctx context.Context, id string, price float64) error {
	event := CarPriceEvent{CarID: id, Price: price}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.nc.Publish("inventory.car.pricing_updated", data)
}
