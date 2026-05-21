package ap2final_test

import (
	"testing"

	bookingv1 "github.com/B4rt0n1/final_proto/gen/go/booking/v1"
	inventoryv1 "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
)

func TestGeneratedContractsAreAvailable(t *testing.T) {
	contractNames := []string{
		string((&bookingv1.Booking{}).ProtoReflect().Descriptor().FullName()),
		string((&inventoryv1.Car{}).ProtoReflect().Descriptor().FullName()),
		string((&userv1.User{}).ProtoReflect().Descriptor().FullName()),
	}

	for _, name := range contractNames {
		if name == "" {
			t.Fatal("generated contract descriptor is empty")
		}
	}
}
