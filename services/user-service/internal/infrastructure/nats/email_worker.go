package nats

import (
	"log"

	"github.com/nats-io/nats.go"
)

type EmailWorker struct {
	nc *nats.Conn
}

func NewEmailWorker(nc *nats.Conn) *EmailWorker {
	return &EmailWorker{nc: nc}
}

func (w *EmailWorker) Start() {
	_, err := w.nc.Subscribe("booking.created", func(m *nats.Msg) {
		log.Printf("Received booking event. Sending email... Payload: %s", string(m.Data))
		// TODO: Интеграция с SMTP
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}
	log.Println("Email worker started, listening to events...")
}
