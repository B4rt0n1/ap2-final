package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
	"github.com/nats-io/nats.go"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type EmailWorker struct {
	nc     *nats.Conn
	users  UserReader
	config SMTPConfig
	send   smtpSender
}

func NewEmailWorker(nc *nats.Conn, users UserReader, config SMTPConfig) *EmailWorker {
	return &EmailWorker{nc: nc, users: users, config: config, send: smtp.SendMail}
}

type smtpSender func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error

func (w *EmailWorker) Start() {
	_, err := w.nc.Subscribe("booking.created", func(message *nats.Msg) {
		if err := w.sendBookingCreatedEmail(message.Data); err != nil {
			log.Printf("booking created email failed: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("failed to subscribe to booking events: %v", err)
	}

	log.Println("email worker started, listening to booking.created")
}

type bookingCreatedEvent struct {
	Booking struct {
		ID         string  `json:"id"`
		UserID     string  `json:"user_id"`
		CarID      string  `json:"car_id"`
		TotalPrice float64 `json:"total_price"`
	} `json:"booking"`
}

func (w *EmailWorker) sendBookingCreatedEmail(payload []byte) error {
	if !w.smtpConfigured() {
		log.Printf("SMTP is not configured, skipping booking email. Payload: %s", string(payload))
		return nil
	}

	var event bookingCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode booking event: %w", err)
	}
	if event.Booking.UserID == "" {
		return fmt.Errorf("booking event has no user_id")
	}

	user, err := w.users.GetByID(context.Background(), event.Booking.UserID)
	if err != nil {
		return fmt.Errorf("load booking user: %w", err)
	}

	subject := "Car rental booking created"
	body := fmt.Sprintf(
		"Your booking %s for car %s was created. Total price: %.2f.",
		event.Booking.ID,
		event.Booking.CarID,
		event.Booking.TotalPrice,
	)
	message := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		user.Email,
		subject,
		body,
	))

	address := w.config.Host + ":" + w.config.Port
	auth := smtp.PlainAuth("", w.config.Username, w.config.Password, w.config.Host)
	if err := w.send(address, auth, w.config.From, []string{user.Email}, message); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}

	log.Printf("booking created email sent to %s", user.Email)
	return nil
}

func (w *EmailWorker) smtpConfigured() bool {
	return w.config.Host != "" &&
		w.config.Port != "" &&
		w.config.Username != "" &&
		w.config.Password != "" &&
		w.config.From != ""
}
