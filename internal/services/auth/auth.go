package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	credentialsFilename  = "credentials.json"
	passwordInitFilename = "password_init"

	hashFilePerm = 0o600
	dirPerm      = 0o700
)

var validUserRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateUser(user string) error {
	user = strings.TrimSpace(user)
	if user == "" {
		return errors.New("user cannot be empty")
	}
	if !validUserRegex.MatchString(user) {
		return errors.New("user contains invalid characters")
	}
	return nil
}

type credentials struct {
	User      string    `json:"user"`
	Hash      string    `json:"hash"`
	ChangedAt time.Time `json:"changed_at"`
}

type NewAuthServiceParams struct {
	HostUser string

	Dir string

	BcryptCost int
}

type AuthService struct {
	mu   sync.RWMutex
	user string
	dir  string
	cost int

	creds *credentials
}

func NewAuthService(params *NewAuthServiceParams) (*AuthService, error) {
	if params == nil {
		params = new(NewAuthServiceParams)
	}
	if params.Dir == "" {
		return nil, fmt.Errorf("auth: Dir is required")
	}
	if err := validateUser(params.HostUser); err != nil {
		return nil, fmt.Errorf("auth: invalid HostUser")
	}

	cost := params.BcryptCost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	s := &AuthService{
		user: params.HostUser,
		dir:  params.Dir,
		cost: cost,
	}

	if err := s.bootstrap(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *AuthService) hashPath() string {
	return filepath.Join(s.dir, credentialsFilename)
}

func (s *AuthService) initPath() string {
	return filepath.Join(s.dir, passwordInitFilename)
}

func (s *AuthService) bootstrap() error {
	exists, err := fileExists(s.initPath())
	if err != nil {
		return fmt.Errorf("auth: failed to check password_init: %w", err)
	}

	if exists {
		return s.consumeInitFile()
	}

	return s.loadFromDisk()
}

func (s *AuthService) consumeInitFile() error {
	raw, err := os.ReadFile(s.initPath())
	if err != nil {
		return fmt.Errorf("auth: failed to read password_init: %w", err)
	}

	password := strings.TrimSpace(string(raw))
	if password == "" {
		_ = os.Remove(s.initPath())
		return fmt.Errorf("auth: password_init is empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.applyNewPasswordLocked(password); err != nil {
		return err
	}

	if err := os.Remove(s.initPath()); err != nil {
		return fmt.Errorf("auth: failed to remove password_init after consuming it: %w", err)
	}

	return nil
}

func (s *AuthService) loadFromDisk() error {
	exists, err := fileExists(s.hashPath())
	if err != nil {
		return fmt.Errorf("auth: failed to check password_hash: %w", err)
	}
	if !exists {
		return nil
	}

	data, err := os.ReadFile(s.hashPath())
	if err != nil {
		return fmt.Errorf("auth: failed to read password_hash: %w", err)
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return fmt.Errorf("%w: %v", ErrCorruptCredentials, err)
	}
	if creds.Hash == "" {
		return fmt.Errorf("%w: empty hash field", ErrCorruptCredentials)
	}

	if err := validateUser(creds.User); err != nil {
		return fmt.Errorf("%w: invalid user field: %v", ErrCorruptCredentials, err)
	}

	s.mu.Lock()
	s.creds = &creds
	s.mu.Unlock()

	return nil
}

func (s *AuthService) applyNewPasswordLocked(password string) error {
	if err := validateUser(s.user); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return fmt.Errorf("auth: failed to hash password: %w", err)
	}

	creds := &credentials{
		User:      s.user,
		Hash:      string(hash),
		ChangedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := s.persist(creds); err != nil {
		return err
	}

	s.creds = creds
	return nil
}

func (s *AuthService) persist(creds *credentials) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("auth: failed to encode credentials: %w", err)
	}

	if err := os.MkdirAll(s.dir, dirPerm); err != nil {
		return fmt.Errorf("auth: failed to create config dir: %w", err)
	}

	tmp, err := os.CreateTemp(s.dir, ".password_hash.*.tmp")
	if err != nil {
		return fmt.Errorf("auth: failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("auth: failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("auth: failed to sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("auth: failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, hashFilePerm); err != nil {
		return fmt.Errorf("auth: failed to set permissions: %w", err)
	}
	if err := os.Rename(tmpPath, s.hashPath()); err != nil {
		return fmt.Errorf("auth: failed to rename into place: %w", err)
	}

	return nil
}

func (s *AuthService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.creds != nil
}

func (s *AuthService) Verify(user, password string) error {
	s.mu.RLock()
	creds := s.creds
	s.mu.RUnlock()

	if creds == nil {
		return s.bootstrapFirstLogin(user, password)
	}

	return compareCredentials(creds, user, password)
}

// bootstrapFirstLogin handles authentication when no credentials have been
// configured yet. The first successful attempt against the configured host
// user permanently sets the supplied password as the admin credential, which
// lets a fresh installation be bootstrapped without a separate "set password"
// step. If credentials are configured concurrently by another request before
// this one acquires the lock, it falls back to verifying against them instead
// of overwriting them.
func (s *AuthService) bootstrapFirstLogin(user, password string) error {
	if user != s.user {
		return ErrInvalidUser
	}
	if strings.TrimSpace(password) == "" {
		return ErrInvalidPassword
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.creds != nil {
		return compareCredentials(s.creds, user, password)
	}

	return s.applyNewPasswordLocked(password)
}

func compareCredentials(creds *credentials, user, password string) error {
	if creds.User != user {
		return ErrInvalidUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.Hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("auth: failed to compare password: %w", err)
	}

	return nil
}

func (s *AuthService) ChangedAt() (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.creds == nil {
		return time.Time{}, ErrNotConfigured
	}
	return s.creds.ChangedAt, nil
}

func (s *AuthService) SetPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("auth: password cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.applyNewPasswordLocked(password)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
