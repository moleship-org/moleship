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
	UsersStrategyFlag string
	UserRepo          port.UserRepository
	SessionRepo       port.SessionRepository
	PasswordManager   port.PasswordManager
	TokenGenerator    port.TokenGenerator
}

type AuthService struct {
	usersStrategyFlag string
	userRepo          port.UserRepository
	sessionRepo       port.SessionRepository
	passwordManager   port.PasswordManager
	tokenGenerator    port.TokenGenerator
}

func NewAuthService(params *AuthServiceParams) *AuthService {
	return &AuthService{
		usersStrategyFlag: params.UsersStrategyFlag,
		userRepo:          params.UserRepo,
		sessionRepo:       params.SessionRepo,
		passwordManager:   params.PasswordManager,
		tokenGenerator:    params.TokenGenerator,
	}
}

// Login authenticates a user with username and password
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	if s.IsOpen() {
		return "", fmt.Errorf("authentication is disabled in open strategy")
	}

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
	token, tokenHash, err := s.tokenGenerator.Generate()
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
	if s.IsOpen() {
		return "", fmt.Errorf("registration is disabled in open strategy")
	}

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

	isAdmin := false
	if s.usersStrategyFlag == "owner_only" {
		// Check if there are any users in the system
		count, err := s.userRepo.Count(ctx)
		if err != nil {
			return "", fmt.Errorf("error counting users: %w", err)
		}

		// If no users exist, make this user the owner (admin)
		if count == 0 {
			isAdmin = true
		}
	}
	if !isAdmin && s.usersStrategyFlag != "multi_user" {
		return "", fmt.Errorf("registration is disabled in '%s' strategy, an owner or admin must register new users", s.usersStrategyFlag)
	}

	// Create user
	user := &model.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		IsAdmin:      isAdmin,
		IsActive:     isAdmin, // New users are inactive by default. The owner/admin can activate them.
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return "", fmt.Errorf("error saving user: %w", err)
	}

	// Create session
	token, tokenHash, err := s.tokenGenerator.Generate()
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
	if s.IsOpen() {
		return "", fmt.Errorf("authentication is disabled in open strategy")
	}

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
	newToken, newTokenHash, err := s.tokenGenerator.Generate()
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
	if s.IsOpen() {
		return fmt.Errorf("authentication is disabled in open strategy")
	}

	// Hash the provided token
	tokenHash := sha256.Sum256([]byte(token))

	// Delete session
	if err := s.sessionRepo.Delete(ctx, tokenHash[:]); err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, error) {
	if s.IsOpen() {
		return "", fmt.Errorf("authentication is disabled in open strategy")
	}

	// Hash the provided token
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

	return session.UserID.String(), nil
}

func (s *AuthService) IsOpen() bool {
	return s.usersStrategyFlag == "open"
}
