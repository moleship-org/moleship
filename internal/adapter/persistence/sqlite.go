package persistence

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/moleship-org/moleship/internal/adapter/db"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func RunMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	log.Println("Verificando migraciones de base de datos...")
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("error al ejecutar migraciones: %w", err)
	}

	log.Println("Base de datos actualizada correctamente.")
	return nil
}

type Repository interface {
	Querier() db.Querier
	DB() *sql.DB
}

type SQLiteRepository struct {
	q *db.Queries
	d *sql.DB
}

func NewSQLiteRepository(conn *sql.DB) *SQLiteRepository {
	sr := new(SQLiteRepository)
	sr.q = db.New(conn)
	sr.d = conn
	return sr
}

func (sr *SQLiteRepository) Querier() db.Querier {
	return sr.q
}

func (sr *SQLiteRepository) DB() *sql.DB {
	return sr.d
}
