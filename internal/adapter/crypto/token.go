package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	DefaultTokenLength = 32 // bytes
)

// TokenGenerator implements the port.TokenGenerator interface
// using crypto/rand for token generation and SHA256 for hashing
type TokenGenerator struct {
	tokenLength int
}

// NewTokenGenerator creates a new TokenGenerator with the default token length
func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{
		tokenLength: DefaultTokenLength,
	}
}

// NewTokenGeneratorWithLength creates a new TokenGenerator with a custom token length
func NewTokenGeneratorWithLength(length int) *TokenGenerator {
	return &TokenGenerator{
		tokenLength: length,
	}
}

// Generate generates a new random token and returns both the token and its hash
func (tg *TokenGenerator) Generate() (token string, hash []byte, err error) {
	randomBytes := make([]byte, tg.tokenLength)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", nil, fmt.Errorf("error generating random bytes: %w", err)
	}

	// Encode token as base64 URL-safe string
	token = base64.URLEncoding.EncodeToString(randomBytes)

	// Hash the token
	hashedToken := sha256.Sum256([]byte(token))
	hash = hashedToken[:]

	return token, hash, nil
}
