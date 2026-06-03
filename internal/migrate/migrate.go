package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run применяет все доступные миграции из указанной папки
func Run(db *sql.DB, logger *slog.Logger, path string) error {
	logger.Info("starting database migrations", slog.String("path", path))

	// Создаём драйвер для postgres на основе существующего подключения
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Инициализируем migrator
	// Важно: путь к файлам должен быть в формате file://путь
	// Для относительного пути: file://migrations
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("migrations applied successfully")
	return nil
}
