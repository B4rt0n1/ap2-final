package nats

import (
	"context"
	"net/smtp"
	"strings"
	"testing"

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/domain"
)

func TestSendBookingCreatedEmailUsesSMTPRecipient(t *testing.T) {
	users := emailUsers{user: &domain.User{ID: "user-1", Email: "renter@example.com"}}
	worker := NewEmailWorker(nil, users, SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "sender",
		Password: "secret",
		From:     "sender@example.com",
	})

	var gotAddress, gotFrom, gotMessage string
	var gotTo []string
	worker.send = func(address string, _ smtp.Auth, from string, to []string, message []byte) error {
		gotAddress = address
		gotFrom = from
		gotTo = append([]string(nil), to...)
		gotMessage = string(message)
		return nil
	}

	payload := []byte(`{"booking":{"id":"booking-1","user_id":"user-1","car_id":"car-1","total_price":255}}`)
	if err := worker.sendBookingCreatedEmail(payload); err != nil {
		t.Fatalf("sendBookingCreatedEmail() error = %v", err)
	}
	if gotAddress != "smtp.example.com:587" || gotFrom != "sender@example.com" {
		t.Fatalf("SMTP target address = %q, from = %q", gotAddress, gotFrom)
	}
	if len(gotTo) != 1 || gotTo[0] != "renter@example.com" {
		t.Fatalf("SMTP recipients = %#v", gotTo)
	}
	if !strings.Contains(gotMessage, "booking-1") || !strings.Contains(gotMessage, "car-1") {
		t.Fatalf("SMTP message = %q", gotMessage)
	}
}

type emailUsers struct {
	user *domain.User
}

func (u emailUsers) GetByID(context.Context, string) (*domain.User, error) {
	return u.user, nil
}
