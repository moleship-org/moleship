package model

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/moleship-org/moleship/internal/adapter/db"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	PasswordHash string     `json:"password_hash"`
	Email        string     `json:"email"`
	IsAdmin      bool       `json:"is_admin"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (u *User) Map(row *db.User) {
	u.ID = uuid.Must(uuid.ParseBytes(row.ID))
	u.Username = row.Username
	u.FirstName = row.FirstName
	u.LastName = row.LastName
	u.PasswordHash = row.PasswordHash
	u.Email = row.Email
	u.IsAdmin = row.IsAdmin
	u.IsActive = row.IsActive

	if row.LastLogin != nil {
		t, err := time.Parse(SQLiteTimeLayout, *row.LastLogin)
		if err != nil {
			log.Fatalf("Error on time.Parse of user LastLogin: %s\n", err.Error())
		}
		u.LastLogin = &t
	}

	t, err := time.Parse(SQLiteTimeLayout, row.CreatedAt)
	if err != nil {
		log.Fatalf("Error on time.Parse of user CreatedAt: %s\n", err.Error())
	}
	u.CreatedAt = t

	t, err = time.Parse(SQLiteTimeLayout, row.UpdatedAt)
	if err != nil {
		log.Fatalf("Error on time.Parse of user UpdatedAt: %s\n", err.Error())
	}
	u.UpdatedAt = t

	if row.DeletedAt != nil {
		t, err := time.Parse(SQLiteTimeLayout, *row.DeletedAt)
		if err != nil {
			log.Fatalf("Error on time.Parse of user DeletedAt: %s\n", err.Error())
		}
		u.DeletedAt = &t
	}
}
