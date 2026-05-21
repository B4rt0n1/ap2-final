package config

import (
	"fmt"
	"os"
)

const (
	defaultGRPCAddress      = ":50053"
	defaultUserAddress      = "localhost:50051"
	defaultInventoryAddress = "localhost:50052"
	defaultNATSURL          = "nats://localhost:4222"
)

// Config contains the booking service runtime endpoints.
type Config struct {
	GRPCAddress      string
	DatabaseURL      string
	UserAddress      string
	InventoryAddress string
	NATSURL          string
}

// Load reads booking runtime configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		GRPCAddress:      envOrDefault("BOOKING_GRPC_ADDRESS", defaultGRPCAddress),
		DatabaseURL:      os.Getenv("BOOKING_DATABASE_URL"),
		UserAddress:      envOrDefault("USER_GRPC_ADDRESS", defaultUserAddress),
		InventoryAddress: envOrDefault("INVENTORY_GRPC_ADDRESS", defaultInventoryAddress),
		NATSURL:          envOrDefault("NATS_URL", defaultNATSURL),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("BOOKING_DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
