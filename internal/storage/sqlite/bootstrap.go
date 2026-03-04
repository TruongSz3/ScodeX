package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const DriverName = "sqlite"

var (
	ErrDatabasePathRequired = errors.New("sqlite: database path is required")
	ErrMigrationIDRequired  = errors.New("sqlite: migration id is required")
)

type Migration struct {
	ID  string
	SQL string
}

func Bootstrap(ctx context.Context, databasePath string, migrations []Migration) (*sql.DB, error) {
	if databasePath == "" {
		return nil, ErrDatabasePathRequired
	}

	db, err := sql.Open(DriverName, databasePath)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping database (modernc driver missing?): %w", err)
	}

	if err := configureWAL(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := applyMigrations(ctx, db, migrations); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func configureWAL(ctx context.Context, db *sql.DB) error {
	queries := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
		fmt.Sprintf("PRAGMA busy_timeout = %d;", (5 * time.Second).Milliseconds()),
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("sqlite: configure pragma %q: %w", query, err)
		}
	}

	return nil
}

func applyMigrations(ctx context.Context, db *sql.DB, migrations []Migration) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("sqlite: ensure schema_migrations: %w", err)
	}

	for _, migration := range migrations {
		if migration.ID == "" {
			return ErrMigrationIDRequired
		}

		applied, err := migrationExists(ctx, db, migration.ID)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("sqlite: start migration tx %q: %w", migration.ID, err)
		}

		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("sqlite: run migration %q: %w", migration.ID, err)
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations(id, applied_at) VALUES(?, ?)",
			migration.ID,
			time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("sqlite: mark migration %q: %w", migration.ID, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("sqlite: commit migration %q: %w", migration.ID, err)
		}
	}

	return nil
}

func migrationExists(ctx context.Context, db *sql.DB, migrationID string) (bool, error) {
	row := db.QueryRowContext(ctx, "SELECT 1 FROM schema_migrations WHERE id = ?", migrationID)
	var marker int
	err := row.Scan(&marker)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("sqlite: read migration %q: %w", migrationID, err)
	}
	return true, nil
}
