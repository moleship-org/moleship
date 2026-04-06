package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moleship-org/moleship/internal/domain/model"
	"github.com/moleship-org/moleship/internal/domain/port"
)

var (
	ErrInvalidCredentials = errors.New("invalidUsername or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
)

type AuthServiceParams struct {
	UserRepo        port.UserRepository
	SessionRepo     port.SessionRepository
	PasswordManager port.PasswordManager
	TokenGenerator  port.TokenGenerator
}

type AuthService struct {
	userRepo        port.UserRepository
	sessionRepo     port.SessionRepository
	passwordManager port.PasswordManager
	tokenGenerator  port.TokenGenerator
}

func NewAuthService(params *AuthServiceParams) *AuthService {
	return &AuthService{
		userRepo:        params.UserRepo,
		sessionRepo:     params.SessionRepo,
		passwordManager: params.PasswordManager,
		tokenGenerator:  params.TokenGenerator,
	}
}

// Login authenticates a user with username and password
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCredentials, err)
	}

	// Check if user is active
	if !user.IsActive {
		return "", fmt.Errorf("user is not active")
	}

	// Compare passwords
	valid, err := s.passwordManager.Compare(user.PasswordHash, password)
	if err != nil {
		return "", fmt.Errorf("error comparing passwords: %w", err)
	}

	if !valid {
		return "", ErrInvalidCredentials
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID.String()); err != nil {
		return "", fmt.Errorf("error updating last login: %w", err)
	}

	// Create session and generate token
	token, tokenHash, err := s.tokenGenerator.GenerateWithExpiry(24 * 60) // 24 hours
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	session := &model.Session{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return "", fmt.Errorf("error saving session: %w", err)
	}

	return token, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, username, email, password string) (string, error) {
	// Check if user already exists
	_, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return "", ErrUserExists
	}

	_, err = s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return "", fmt.Errorf("email already registered")
	}

	// Hash password
	passwordHash, err := s.passwordManager.Hash(password)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &model.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		IsAdmin:      false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return "", fmt.Errorf("error saving user: %w", err)
	}

	// Create session
	token, tokenHash, err := s.tokenGenerator.GenerateWithExpiry(24 * 60) // 24 hours
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	session := &model.Session{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return "", fmt.Errorf("error saving session: %w", err)
	}

	return token, nil
}

// Refresh generates a new token for an existing valid session
func (s *AuthService) Refresh(ctx context.Context, token string) (string, error) {
	// Hash the provided token to look it up
	tokenHash := sha256.Sum256([]byte(token))

	// Find session by token hash
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash[:])
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		if err := s.sessionRepo.Delete(ctx, tokenHash[:]); err != nil {
			return "", fmt.Errorf("error deleting expired session: %w", err)
		}
		return "", ErrInvalidToken
	}

	// Generate new token
	newToken, newTokenHash, err := s.tokenGenerator.GenerateWithExpiry(24 * 60) // 24 hours
	if err != nil {
		return "", fmt.Errorf("error generating new token: %w", err)
	}

	// Delete old session
	if err := s.sessionRepo.Delete(ctx, tokenHash[:]); err != nil {
		return "", fmt.Errorf("error deleting old session: %w", err)
	}

	// Create new session
	newSession := &model.Session{
		TokenHash: newTokenHash,
		UserID:    session.UserID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Save(ctx, newSession); err != nil {
		return "", fmt.Errorf("error saving new session: %w", err)
	}

	return newToken, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(ctx context.Context, token string) error {
	// Hash the provided token
	tokenHash := sha256.Sum256([]byte(token))

	// Delete session
	if err := s.sessionRepo.Delete(ctx, tokenHash[:]); err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}
