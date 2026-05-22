package config

import (
	"fmt"
	"os"
)

const (
	defaultGRPCAddress = ":50051"
	defaultNATSURL     = "nats://localhost:4222"
)

type Config struct {
	GRPCAddress string
	DatabaseURL string
	NATSURL     string
	SMTP        SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func Load() (Config, error) {
	cfg := Config{
		GRPCAddress: envOrDefault("USER_GRPC_ADDRESS", defaultGRPCAddress),
		DatabaseURL: os.Getenv("USER_DATABASE_URL"),
		NATSURL:     envOrDefault("NATS_URL", defaultNATSURL),
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("USER_DATABASE_URL is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
