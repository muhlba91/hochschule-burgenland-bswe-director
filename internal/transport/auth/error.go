package auth

import (
	"errors"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
)

var (
	// ErrSecretNotConfigured is returned when the authentication secret is not configured.
	ErrSecretNotConfigured = errors.New("secret not configured")
	// ErrUnexpectedSigningMethod is returned when the JWT signing method is not as expected.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	// ErrInvalidToken is returned when the JWT token is invalid or expired.
	ErrInvalidToken = errors.New("invalid token")
	// ErrMissingToken is returned when the authentication token is missing.
	ErrMissingToken = errors.New("missing authentication token")
	// ErrInvalidRequest is returned when the request payload is invalid.
	ErrInvalidRequest = response.ErrInvalidPayload
	// ErrInvalidCredentials is returned when the provided API key or client ID is invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
