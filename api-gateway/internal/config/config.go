package config

import "os"

const (
	defaultHTTPAddress     = ":8080"
	defaultBookingTarget   = "localhost:50053"
	defaultUserTarget      = "localhost:50052" // adjust as needed
	defaultInventoryTarget = "localhost:50051" // inventory service port
)

type Config struct {
	HTTPAddress     string
	BookingTarget   string
	UserTarget      string
	InventoryTarget string
}

func Load() Config {
	return Config{
		HTTPAddress:     envOrDefault("GATEWAY_HTTP_ADDRESS", defaultHTTPAddress),
		BookingTarget:   envOrDefault("BOOKING_GRPC_TARGET", defaultBookingTarget),
		UserTarget:      envOrDefault("USER_GRPC_TARGET", defaultUserTarget),
		InventoryTarget: envOrDefault("INVENTORY_GRPC_TARGET", defaultInventoryTarget),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
