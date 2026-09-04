package auth

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/moleship-org/moleship/internal/domain/crypto"
	"github.com/moleship-org/moleship/internal/domain/persistence"
	_ "modernc.org/sqlite"
)

func TestLoginSucceedsAfterRegisterWithSQLiteTextTimestamps(t *testing.T) {
	dir, err := os.MkdirTemp("", "moleship-auth-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "moleship.db")
	conn, err := sql.Open("sqlite", "file:"+dbPath+"?cache=shared&_fk=1")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := persistence.RunMigrations(conn, "../../../database/migrations"); err != nil {
		t.Fatal(err)
	}

	repo := persistence.NewSQLiteRepository(conn)
	userRepo := persistence.NewUserRepository(repo)
	sessionRepo := persistence.NewSessionRepository(repo)
	service := NewAuthService(&AuthServiceParams{
		UsersStrategyFlag: "multi_user",
		UserRepo:          userRepo,
		SessionRepo:       sessionRepo,
		PasswordManager:   crypto.NewDefaultPasswordManager(),
		TokenGenerator:    crypto.NewTokenGenerator(),
	})

	if _, err := service.Register(context.Background(), "alice", "alice@example.com", "password123"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	user, err := userRepo.FindByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if user == nil || user.Username != "alice" {
		t.Fatal("expected user to be returned from FindByUsername")
	}

	if _, err := service.Login(context.Background(), "alice", "password123"); err != nil {
		t.Fatalf("login failed: %v", err)
	}
}
