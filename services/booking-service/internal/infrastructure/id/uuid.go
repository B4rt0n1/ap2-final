package id

import "github.com/google/uuid"

// UUIDGenerator creates random booking IDs.
type UUIDGenerator struct{}

// NewID returns a random UUID string.
func (UUIDGenerator) NewID() string {
	return uuid.NewString()
}
