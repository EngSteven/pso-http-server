package util

import "github.com/google/uuid"

// NewRequestID genera un identificador único para cada request.
func NewRequestID() string {
	return uuid.NewString()
}
