package natspublisher

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/usecase"
)

func TestPublishBookingEventUsesEventSubjectAndJSONPayload(t *testing.T) {
	connection := &recordingConnection{}
	publisher := &Publisher{connection: connection}

	err := publisher.PublishBookingEvent(context.Background(), usecase.BookingEvent{
		Type:       usecase.EventBookingConfirmed,
		OccurredAt: time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		Booking: usecase.BookingSnapshot{
			ID:     "booking-1",
			UserID: "user-1",
			CarID:  "car-1",
			Status: "confirmed",
		},
	})
	if err != nil {
		t.Fatalf("PublishBookingEvent() error = %v", err)
	}

	if connection.subject != "booking.confirmed" {
		t.Fatalf("PublishBookingEvent() subject = %q, want booking.confirmed", connection.subject)
	}

	var payload usecase.BookingEvent
	if err := json.Unmarshal(connection.data, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Booking.ID != "booking-1" {
		t.Fatalf("PublishBookingEvent() booking id = %q, want booking-1", payload.Booking.ID)
	}
}

type recordingConnection struct {
	subject string
	data    []byte
}

func (c *recordingConnection) Publish(subject string, data []byte) error {
	c.subject = subject
	c.data = data
	return nil
}
