package config

import "os"

const (
	defaultHTTPAddress   = ":8080"
	defaultBookingTarget = "localhost:50053"
)

// Config contains API Gateway runtime endpoints.
type Config struct {
	HTTPAddress   string
	BookingTarget string
}

// Load reads API Gateway configuration from environment variables.
func Load() Config {
	return Config{
		HTTPAddress:   envOrDefault("GATEWAY_HTTP_ADDRESS", defaultHTTPAddress),
		BookingTarget: envOrDefault("BOOKING_GRPC_TARGET", defaultBookingTarget),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
