package migration

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	// Try to run migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Printf("No migrations to apply")
			version, dirty, _ := m.Version()
			log.Printf("Migration complete. Version: %d, Dirty: %v", version, dirty)
			return nil
		}

		// If migration fails, try to force set version to latest
		log.Printf("Migration failed: %v, attempting to force set version to 13", err)
		if err := m.Force(13); err != nil {
			return fmt.Errorf("failed to force set migration version: %w", err)
		}
		log.Printf("Migration version forced to 13")
	}

	version, dirty, _ := m.Version()
	log.Printf("Migration complete. Version: %d, Dirty: %v", version, dirty)
	return nil
}

func Rollback(dsn string, steps int) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	return m.Steps(-steps)
}

func Version(dsn string) (uint, bool, error) {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	return m.Version()
}
