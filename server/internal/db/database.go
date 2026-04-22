package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.up.sql
var migrationFS embed.FS

// Config holds database connection parameters.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Database wraps a sql.DB with migration support.
type Database struct {
	*sql.DB
}

// New creates a new Database connection and runs migrations.
func New(ctx context.Context, cfg Config) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{DB: db}

	if err := d.RunMigrations(ctx); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return d, nil
}

// RunMigrations applies all pending migration files in order.
func (d *Database) RunMigrations(ctx context.Context) error {
	if err := d.createSchemaTable(ctx); err != nil {
		return err
	}

	migrations, err := d.loadMigrations()
	if err != nil {
		return err
	}

	applied, err := d.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.name] {
			continue
		}

		if err := d.applyMigration(ctx, m); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.name, err)
		}
		applied[m.name] = true
	}

	return nil
}

type migration struct {
	name    string
	order   int
	content string
}

func (d *Database) loadMigrations() ([]migration, error) {
	var migrations []migration

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		order := 0
		parts := strings.SplitN(name, "_", 2)
		if len(parts) > 0 && parts[0] != "" {
			order, _ = strconv.Atoi(parts[0])
		}

		migrations = append(migrations, migration{
			name:    name,
			order:   order,
			content: string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].order < migrations[j].order
	})

	return migrations, nil
}

func (d *Database) createSchemaTable(ctx context.Context) error {
	_, err := d.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	return err
}

func (d *Database) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	applied := make(map[string]bool)
	rows, err := d.QueryContext(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}
	return applied, rows.Err()
}

func (d *Database) applyMigration(ctx context.Context, m migration) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, m.content); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", m.name); err != nil {
		return err
	}

	return tx.Commit()
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.DB.Close()
}

// ColumnExists checks whether a column exists in a table.
func (d *Database) ColumnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	var count int
	err := d.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_name = $1 AND column_name = $2
	`, tableName, columnName).Scan(&count)
	return count > 0, err
}

// ReadMigrationFiles returns all migration file names and content for testing.
func ReadMigrationFiles() (map[string]string, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, f := range entries {
		if f.IsDir() {
			continue
		}
		content, err := migrationFS.ReadFile("migrations/" + f.Name())
		if err != nil {
			return nil, err
		}
		result[f.Name()] = string(content)
	}
	return result, nil
}

// MigrationCount returns the number of migration files.
func MigrationCount() (int, error) {
	files, err := ReadMigrationFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// ValidateMigrations checks that migration files are well-formed.
func ValidateMigrations() error {
	files, err := ReadMigrationFiles()
	if err != nil {
		return err
	}
	for name, content := range files {
		if len(strings.TrimSpace(content)) == 0 {
			return fmt.Errorf("migration %s is empty", name)
		}
	}
	return nil
}
