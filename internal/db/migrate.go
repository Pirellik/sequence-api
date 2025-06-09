package db

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsDir embed.FS

func MigrateUp(url string) error {
	dir, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("creating iofs for migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", dir, url)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err.Error() != "no change" {
		return fmt.Errorf("migrating up: %w", err)
	}

	return nil
}
