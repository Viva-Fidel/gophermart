package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var migrationUpPattern = regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)

type migration struct {
	version int
	name    string
	path    string
}

// Применяет все новые миграции к базе данных
func RunMigrations(ctx context.Context, db *sql.DB, dir string) error {
	// Создаёт таблицу schema_migrations, если её нет
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}

	// Загружает и сортирует .up.sql миграции из указанной директории
	files, err := loadUpMigrations(dir)
	if err != nil {
		return err
	}

	// Применяет каждую миграцию
	for _, m := range files {
		// Проверяет, не была ли миграция уже применена
		var alreadyApplied bool
		if err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			m.version,
		).Scan(&alreadyApplied); err != nil {
			return fmt.Errorf("failed to check migration %d: %w", m.version, err)
		}
		if alreadyApplied {
			continue
		}

		// Читает файл миграции
		sqlBytes, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("failed to read migration %q: %w", m.path, err)
		}

		// Начинает транзакцию
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start tx for migration %d: %w", m.version, err)
		}

		// Выполняет миграцию
		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "migration rollback after exec error", slog.Int("version", m.version), slog.Any("rollback_error", rbErr), slog.Any("exec_error", err))
			}
			return fmt.Errorf("failed to apply migration %d: %w", m.version, err)
		}

		// Сохраняет версию миграции
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(version, name) VALUES ($1, $2)`,
			m.version,
			m.name,
		); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "migration rollback after version insert error", slog.Int("version", m.version), slog.Any("rollback_error", rbErr), slog.Any("exec_error", err))
			}
			return fmt.Errorf("failed to save migration version %d: %w", m.version, err)
		}

		// Фиксирует транзакцию
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.version, err)
		}
	}

	return nil
}

// Создаёт таблицу schema_migrations, если её нет
func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	// Создаёт таблицу schema_migrations, если её нет
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}
	return nil
}

// Загружает и сортирует .up.sql миграции из указанной директории
func loadUpMigrations(dir string) ([]migration, error) {
	// Читает директорию с миграциями
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir %q: %w", dir, err)
	}

	// Создаёт слайс миграций
	migrations := make([]migration, 0, len(entries))
	// Проходит по всем файлам в директории
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		match := migrationUpPattern.FindStringSubmatch(name)
		if len(match) != 2 {
			continue
		}
		version, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("bad migration version in %q: %w", name, err)
		}
		// Добавляет миграцию в слайс
		migrations = append(migrations, migration{
			version: version,
			name:    name,
			path:    filepath.Join(dir, name),
		})
	}

	// Сортирует миграции по версии
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}
