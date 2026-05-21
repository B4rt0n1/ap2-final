package natspublisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/usecase"
	"github.com/nats-io/nats.go"
)

var _ usecase.EventPublisher = (*Publisher)(nil)

type connection interface {
	Publish(subject string, data []byte) error
}

// Publisher sends booking lifecycle JSON events to NATS subjects.
type Publisher struct {
	connection connection
}

// New creates a core NATS booking event publisher.
func New(connection *nats.Conn) *Publisher {
	return &Publisher{connection: connection}
}

// PublishBookingEvent publishes one booking lifecycle fact to its event subject.
func (p *Publisher) PublishBookingEvent(ctx context.Context, event usecase.BookingEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal booking event: %w", err)
	}

	if err := p.connection.Publish(string(event.Type), payload); err != nil {
		return fmt.Errorf("publish booking event: %w", err)
	}

	return nil
}
