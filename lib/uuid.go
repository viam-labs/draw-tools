package lib

import (
	"github.com/google/uuid"
)

// UUID is a wrapper around github.com/google/uuid.UUID providing convenient methods.
// It encapsulates UUID functionality with methods for string and byte conversion.
type UUID struct {
	data uuid.UUID
}

// String returns the string representation of the UUID.
// Returns the UUID in standard format (e.g., "550e8400-e29b-41d4-a716-446655440000").
func (u *UUID) String() string {
	return u.data.String()
}

// Bytes returns the byte array representation of the UUID.
// Returns a 16-byte array containing the UUID data.
func (u *UUID) Bytes() []byte {
	return u.data[:]
}

// GenerateUUID generates a new random UUID.
// Creates a version 4 (random) UUID using the underlying uuid.New() function.
//
// Returns a new random UUID.
func GenerateUUID() UUID {
	id := uuid.New()
	return UUID{
		data: id,
	}
}

// UUIDFromBytes creates a UUID from a byte array.
// The byte array must be exactly 16 bytes long and contain valid UUID data.
//
// Parameters:
//   - bytes: 16-byte array containing UUID data
//
// Returns the UUID created from bytes or an error if the bytes are invalid.
func UUIDFromBytes(bytes []byte) (*UUID, error) {
	id, err := uuid.FromBytes(bytes)
	if err != nil {
		return nil, err
	}

	return &UUID{
		data: id,
	}, nil
}

// UUIDFromString creates a UUID from a string representation.
// The string must be in standard UUID format (e.g., "550e8400-e29b-41d4-a716-446655440000").
// If the string is empty, a new random UUID is generated.
//
// Parameters:
//   - id: UUID string in standard format
//
// Returns the UUID created from string or an error if the string format is invalid.
func UUIDFromString(id string) (*UUID, error) {
	if id == "" {
		newUUID := GenerateUUID()
		return &newUUID, nil
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return &UUID{
		data: parsed,
	}, nil
}
