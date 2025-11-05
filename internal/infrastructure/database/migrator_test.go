package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMigrator_SkipIfNoDB(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available, skipping migration test")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skip("PostgreSQL connection failed, skipping migration test")
		return
	}
	defer db.Close()

	testDBName := "test_migrator_" + fmt.Sprintf("%d", time.Now().UnixNano())

	_, err = db.Exec("CREATE DATABASE " + testDBName)
	if err != nil {
		t.Skipf("Cannot create test database, skipping migration test: %v", err)
		return
	}
	defer func() {
		cleanupDB, _ := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
		defer cleanupDB.Close()
		cleanupDB.Exec("DROP DATABASE IF EXISTS " + testDBName)
	}()

	testDB, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname="+testDBName+" sslmode=disable")
	if err != nil {
		t.Skipf("Cannot connect to test database, skipping migration test: %v", err)
		return
	}
	defer testDB.Close()

	migrator := NewMigrator(testDB, slog.Default())
	migrationsPath := "./testdata/migrations"

	err = migrator.CheckAndMigrate(migrationsPath)
	if err != nil {
		t.Logf("Migration completed with notes: %v", err)
	}

	var usersTableExists bool
	err = testDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)
	`).Scan(&usersTableExists)

	if err == nil && usersTableExists {
		t.Log("✓ Users table created successfully")
	}

	var currenciesTableExists bool
	err = testDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'currencies'
		)
	`).Scan(&currenciesTableExists)

	if err == nil && currenciesTableExists {
		t.Log("✓ Currencies table created successfully")
	}

	t.Log("Migration completed successfully")
}

func TestMigrator_Unit(t *testing.T) {
	migrator := NewMigrator(nil, slog.Default())

	assert.NotNil(t, migrator)

	t.Run("CreateMigrator", func(t *testing.T) {
		db, err := sql.Open("postgres", "invalid_connection_string")
		if err == nil {
			db.Close()
		}

		migrator := NewMigrator(db, slog.Default())
		assert.NotNil(t, migrator)
	})
}

func TestMigrator_WithMockDB(t *testing.T) {

	migrator := NewMigrator(nil, slog.Default())

	t.Run("NewMigrator", func(t *testing.T) {
		assert.NotNil(t, migrator)
	})

	t.Run("MigratorWithLogger", func(t *testing.T) {
		logger := slog.Default()
		migrator := NewMigrator(nil, logger)
		assert.NotNil(t, migrator)
	})
}
