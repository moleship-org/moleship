package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = bcrypt.DefaultCost

// PasswordManager implements the port.PasswordManager interface
// using bcrypt for password hashing and comparison
type PasswordManager struct {
	cost int
}

// NewPasswordManager creates a new PasswordManager with the given cost
func NewPasswordManager(cost int) *PasswordManager {
	c := defaultBcryptCost
	if cost > 0 {
		c = cost
	}
	return &PasswordManager{
		cost: c,
	}
}

// NewPasswordManagerWithCost creates a new PasswordManager with a custom bcrypt cost
func NewPasswordManagerWithCost(cost int) *PasswordManager {
	return &PasswordManager{
		cost: cost,
	}
}

// Hash hashes a password using bcrypt
func (pm *PasswordManager) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hash), nil
}

// Compare compares a password hash with a plain password
func (pm *PasswordManager) Compare(passwordHash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("error comparing passwords: %w", err)
	}
	return true, nil
}
