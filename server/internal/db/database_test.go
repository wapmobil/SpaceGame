package db

import (
	"testing"
)

func TestMigrationCount(t *testing.T) {
	count, err := MigrationCount()
	if err != nil {
		t.Fatalf("failed to get migration count: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one migration file")
	}
}

func TestValidateMigrations(t *testing.T) {
	if err := ValidateMigrations(); err != nil {
		t.Fatalf("migration validation failed: %v", err)
	}
}

func TestReadMigrationFiles(t *testing.T) {
	files, err := ReadMigrationFiles()
	if err != nil {
		t.Fatalf("failed to read migration files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one migration file")
	}
	for name, content := range files {
		if len(content) == 0 {
			t.Fatalf("migration %s has empty content", name)
		}
	}
}
