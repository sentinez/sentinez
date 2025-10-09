package migratepgx

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(srcFiles, dbUrl, action string, step int) error {
	m, err := migrate.New(srcFiles, dbUrl)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	defer func() { _, _ = m.Close() }()

	switch action {
	case "up":
		err = m.Up()
	case "down":
		if step <= 0 {
			step = 1
		}
		err = m.Steps(-step)
	default:
		return fmt.Errorf("unknown action: %s (use 'up' or 'down')", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration error: %w", err)
	}
	return nil
}
