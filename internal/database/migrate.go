package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// Migrate runs all pending database migrations.
func Migrate(ctx context.Context, db *sql.DB) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("getting applied migrations: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		log.Printf("applying migration %d: %s", m.version, m.name)
		if err := applyMigration(ctx, db, m); err != nil {
			return fmt.Errorf("applying migration %d: %w", m.version, err)
		}
	}

	return nil
}

type migration struct {
	version int
	name    string
	sql     string
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

func getAppliedMigrations(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func loadMigrations() ([]migration, error) {
	var migrations []migration

	err := fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		filename := filepath.Base(path)
		version, name, err := parseMigrationFilename(filename)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", filename, err)
		}

		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		migrations = append(migrations, migration{
			version: version,
			name:    name,
			sql:     string(content),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func parseMigrationFilename(filename string) (int, string, error) {
	// Expected format: 000001_create_sessions_table.up.sql
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename format")
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version number: %w", err)
	}

	name := strings.TrimSuffix(parts[1], ".up.sql")
	return version, name, nil
}

func applyMigration(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return fmt.Errorf("executing sql: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	return tx.Commit()
}
