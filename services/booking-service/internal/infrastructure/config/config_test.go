package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("BOOKING_DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want required database URL error")
	}
}

func TestLoadUsesDefaultsAndOverrides(t *testing.T) {
	t.Setenv("BOOKING_DATABASE_URL", "postgres://booking")
	t.Setenv("BOOKING_GRPC_ADDRESS", ":6000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.GRPCAddress != ":6000" {
		t.Fatalf("Load() gRPC address = %q, want :6000", cfg.GRPCAddress)
	}
	if cfg.NATSURL != defaultNATSURL {
		t.Fatalf("Load() NATS URL = %q, want %q", cfg.NATSURL, defaultNATSURL)
	}
}
