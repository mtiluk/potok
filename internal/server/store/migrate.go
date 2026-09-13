package store

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/mtiluk/potok/migrations"
)

func (s *Store) migrator() (*migrate.Migrate, error) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("store: read migrations: %w", err)
	}

	driver, err := sqlite.WithInstance(s.db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("store: migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return nil, fmt.Errorf("store: migrator: %w", err)
	}
	return m, nil
}

func (s *Store) Migrate() error {
	m, err := s.migrator()
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("store: apply migrations: %w", err)
	}
	return nil
}

func (s *Store) Version() (version uint, dirty bool, err error) {
	m, err := s.migrator()
	if err != nil {
		return 0, false, err
	}

	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("store: read schema version: %w", err)
	}
	return version, dirty, nil
}
