package provider

import "errors"

var (
	// ErrTokenExpired is used to signal that a token is expired.
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidCredentials is used to signal credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is used to signal that a token is not encoded properly.
	ErrInvalidToken = errors.New("invalid token")
	// ErrInvalidSecretSize is used to signal that an auth provider's secret is
	// not 64 bytes long.
	ErrInvalidSecretSize = errors.New("secret must be 64 bytes long")
)
