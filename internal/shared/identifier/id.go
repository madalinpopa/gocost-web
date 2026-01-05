package identifier

import (
	"encoding/base64"
	"errors"

	"github.com/google/uuid"
)

// ErrInvalidID indicates that the provided ID format is invalid
var ErrInvalidID = errors.New("invalid id format")

// ID represents a stable identifier that can be safely serialized.
type ID struct {
	value uuid.UUID
}

func NewID() (ID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return ID{}, err
	}
	return ID{value: id}, nil
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, ErrInvalidID
	}
	return ID{value: id}, nil
}

func ParseEncodedID(s string) (ID, error) {
	return decode(s)
}

func (i ID) String() string {
	return i.value.String()
}

func (i ID) EncodedString() string {
	return base64.RawURLEncoding.EncodeToString(i.value[:])
}

func (i ID) Equals(other ID) bool {
	return i.value == other.value
}

func (i ID) IsZero() bool {
	return i.value == uuid.Nil
}

func (i ID) UUID() uuid.UUID {
	return i.value
}

func decode(s string) (ID, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(s)
		if err != nil {
			return ID{}, ErrInvalidID
		}
	}
	value, err := uuid.FromBytes(data)
	if err != nil {
		return ID{}, ErrInvalidID
	}
	return ID{value: value}, nil
}
