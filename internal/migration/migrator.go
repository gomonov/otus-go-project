package migrations

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Conf struct {
	AutoMigrate bool
	Dir         string
	Dsn         string
}

func AutoMigrate(logger Logger, cfg Conf) error {
	if !cfg.AutoMigrate {
		logger.Info("Auto migrations disabled")
		return nil
	}

	db, err := sql.Open("postgres", cfg.Dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(fmt.Sprintf("Failed to close database connection: %v", err))
		}
	}()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("Applying migrations...")
	if err := goose.Up(db, cfg.Dir); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Migrations applied successfully")

	return nil
}
