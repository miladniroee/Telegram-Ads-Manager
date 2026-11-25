package database

import (
	"context"
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

// RestoreBackup replaces the current database records with records from a backup.
func RestoreBackup(db *sql.DB, source string) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return fmt.Errorf("please enter a backup file path")
	}

	absolutePath, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return fmt.Errorf("cannot access backup file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("backup path must point to a regular file")
	}

	ctx := context.Background()

	// Use one connection because ATTACH DATABASE is connection-specific.
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prevent restoring the database file onto itself.
	rows, err := conn.QueryContext(ctx, "PRAGMA database_list")
	if err != nil {
		return err
	}

	var currentPath string

	for rows.Next() {
		var sequence int
		var name string
		var file string

		if err := rows.Scan(&sequence, &name, &file); err != nil {
			rows.Close()
			return err
		}

		if name == "main" {
			currentPath = file
			break
		}
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	if currentPath != "" {
		currentAbsolutePath, err := filepath.Abs(currentPath)
		if err != nil {
			return err
		}

		if filepath.Clean(currentAbsolutePath) == filepath.Clean(absolutePath) {
			return fmt.Errorf("cannot restore the database from itself")
		}
	}

	// Attach the backup to this connection.
	_, err = conn.ExecContext(
		ctx,
		"ATTACH DATABASE ? AS restore_source",
		absolutePath,
	)
	if err != nil {
		return fmt.Errorf("could not open backup database: %w", err)
	}
	defer conn.ExecContext(ctx, "DETACH DATABASE restore_source")

	// Check the backup's SQLite integrity.
	integrityRows, err := conn.QueryContext(
		ctx,
		"PRAGMA restore_source.integrity_check",
	)
	if err != nil {
		return err
	}

	integrityOK := true

	for integrityRows.Next() {
		var result string
		if err := integrityRows.Scan(&result); err != nil {
			integrityRows.Close()
			return err
		}

		if result != "ok" {
			integrityOK = false
		}
	}

	if err := integrityRows.Err(); err != nil {
		integrityRows.Close()
		return err
	}
	integrityRows.Close()

	if !integrityOK {
		return fmt.Errorf("backup failed SQLite integrity check")
	}

	// Check that the backup has the expected tables and columns.
	checks := []string{
		`SELECT id, ad_date, ad_time, customer_name, removal_date, price, is_finished
		 FROM restore_source.advertisements LIMIT 0`,
		`SELECT id, amount, payment_date, description, receipt_path
		 FROM restore_source.payments LIMIT 0`,
		`SELECT id, payment_id, advertisement_id, allocated_amount
		 FROM restore_source.payment_allocations LIMIT 0`,
	}

	for _, query := range checks {
		if _, err := conn.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("backup database has an incompatible structure: %w", err)
		}
	}

	// Check foreign-key relationships in the backup.
	fkRows, err := conn.QueryContext(
		ctx,
		"PRAGMA restore_source.foreign_key_check",
	)
	if err != nil {
		return err
	}

	hasForeignKeyErrors := fkRows.Next()
	fkErr := fkRows.Err()
	fkRows.Close()

	if fkErr != nil {
		return fkErr
	}
	if hasForeignKeyErrors {
		return fmt.Errorf("backup contains invalid foreign-key relationships")
	}

	// Replace all records in a single transaction.
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete allocations first because they reference both other tables.
	_, err = tx.ExecContext(ctx, `DELETE FROM payment_allocations`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM payments`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM advertisements`)
	if err != nil {
		return err
	}

	// Restore advertisements.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO advertisements (
			id, ad_date, ad_time, customer_name,
			removal_date, price, is_finished
		)
		SELECT
			id, ad_date, ad_time, customer_name,
			removal_date, price, is_finished
		FROM restore_source.advertisements
	`)
	if err != nil {
		return err
	}

	// Restore payments.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (
			id, amount, payment_date, description, receipt_path
		)
		SELECT
			id, amount, payment_date, description, receipt_path
		FROM restore_source.payments
	`)
	if err != nil {
		return err
	}

	// Restore allocations.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment_allocations (
			id, payment_id, advertisement_id, allocated_amount
		)
		SELECT
			id, payment_id, advertisement_id, allocated_amount
		FROM restore_source.payment_allocations
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}
