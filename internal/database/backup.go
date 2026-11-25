package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateBackup creates a consistent SQLite backup at the given destination.
func CreateBackup(db *sql.DB, destination string) error {
	destination = strings.TrimSpace(destination)

	if destination == "" {
		return fmt.Errorf("please enter a backup file path")
	}

	absolutePath, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	// Find the current database file.
	rows, err := db.Query("PRAGMA database_list")
	if err != nil {
		return err
	}

	var sourcePath string

	for rows.Next() {
		var sequence int
		var name string
		var file string

		if err := rows.Scan(&sequence, &name, &file); err != nil {
			rows.Close()
			return err
		}

		if name == "main" {
			sourcePath = file
			break
		}
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// Prevent backing up over the original database.
	if sourcePath != "" {
		sourceAbsolutePath, err := filepath.Abs(sourcePath)
		if err != nil {
			return err
		}

		if filepath.Clean(sourceAbsolutePath) == filepath.Clean(absolutePath) {
			return fmt.Errorf("backup destination cannot be the original database file")
		}
	}

	// Create the destination directory if necessary.
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0755); err != nil {
		return fmt.Errorf("could not create backup directory: %w", err)
	}

	// Do not overwrite an existing backup.
	if _, err := os.Stat(absolutePath); err == nil {
		return fmt.Errorf("a file already exists at this path")
	} else if !os.IsNotExist(err) {
		return err
	}

	// VACUUM INTO creates a consistent snapshot of the database.
	_, err = db.Exec("VACUUM INTO ?", absolutePath)
	if err != nil {
		return fmt.Errorf("could not create database backup: %w", err)
	}

	return nil
}
