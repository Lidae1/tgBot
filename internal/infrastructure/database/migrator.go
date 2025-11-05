package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMigrator(db *sql.DB, logger *slog.Logger) *Migrator {
	return &Migrator{db: db, logger: logger.With(slog.String("component", "migrator"))}
}

func (m *Migrator) Up(migrationsDir string) error {
	m.logger.Info("Running database migrations", "dir", migrationsDir)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(m.db, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	m.logger.Info("Database migrations completed successfully")
	return nil
}

func (m *Migrator) Down(migrationsDir string) error {
	m.logger.Info("Rolling back database migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Down(m.db, migrationsDir); err != nil {
		return fmt.Errorf("rollback migrations: %w", err)
	}

	m.logger.Info("Database migrations rolled back successfully")
	return nil
}

func (m *Migrator) Status(migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	return goose.Status(m.db, migrationsDir)
}

func (m *Migrator) CheckAndMigrate(migrationsDir string) error {
	if err := m.Status(migrationsDir); err != nil {
		m.logger.Warn("Failed to check migration status", "error", err)
	}

	return m.Up(migrationsDir)
}
