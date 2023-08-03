package spass

import (
	"context"
)

type Pair struct {
	Key string
	Value string
}

type Secret interface {
	// The name of the secret
	Name() string

	// The namespace of the secret, if any.
	Namespace() string

	// The name, including the namespace.
	Fullname() string

	// Get the password in the secret
	Password() (string, error)

	// The full body of the secret
	Body() (string, error)

	// SetPassword sets the password of the secret
	SetPassword(string) error

	// All the value, except for the secret
	Pairs() ([]*Pair, error)

	// Remove the secret
	Remove() error
}

type Store interface {
	// List all secrets in a store.
	// If namespace is not the empty string, it will filter by the provided namespace.
	List(ctx context.Context, namepace string) ([]Secret, error)

	// Get the secret by name.
	Secret(ctx context.Context, name string) (Secret, error)

	// Create a new secret.
	NewSecret(ctx context.Context) (Secret, error)
}
