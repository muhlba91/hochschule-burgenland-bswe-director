package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims.
type Claims struct {
	jwt.RegisteredClaims

	// ClientID is the unique identifier for the client.
	ClientID string `json:"client_id"`
}

// VerifyAPIKey checks if the provided API key is valid for the given client ID and secret.
// clientID: The unique identifier for the client.
// apiKey: The API key provided by the client.
// secret: The secret used to generate the expected API key.
func VerifyAPIKey(clientID string, apiKey string, secret string) bool {
	if secret == "" || clientID == "" || apiKey == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(clientID))
	expectedKey := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(apiKey), []byte(expectedKey))
}

// GenerateJWT creates a new short-lived JWT for a client.
// clientID: The unique identifier for the client.
// secret: The secret used for signing the JWT.
// ttlSeconds: The time-to-live for the JWT in seconds.
func GenerateJWT(clientID string, secret string, ttlSeconds int) (string, error) {
	if secret == "" {
		return "", ErrSecretNotConfigured
	}

	claims := &Claims{
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttlSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT parses and validates a JWT string.
// tokenString: The JWT string to validate.
// secret: The secret used for signing the JWT.
func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
