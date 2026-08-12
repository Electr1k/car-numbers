package postgres

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate накатывает миграции из каталога path на базу по databaseURL
func Migrate(databaseURL, path string) (err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve migrations path %q: %w", path, err)
	}

	migrator, err := migrate.New("file://"+absPath, databaseURL)
	if err != nil {
		return fmt.Errorf("init migrator (source %s): %w", absPath, err)
	}

	defer func() {
		// Close возвращает две ошибки - от источника и от базы.
		sourceErr, dbErr := migrator.Close()
		if err == nil {
			err = errors.Join(sourceErr, dbErr)
		}
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
