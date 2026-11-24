package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "telegram_ads.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, err
	}

	err = createTables(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS advertisements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ad_date TEXT NOT NULL,
		ad_time INTEGER NOT NULL CHECK(ad_time BETWEEN 0 AND 23),
		customer_name TEXT NOT NULL,
		removal_date TEXT NOT NULL,
		price INTEGER NOT NULL CHECK(price >= 0),
		is_finished INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount INTEGER NOT NULL CHECK(amount > 0),
		payment_date TEXT NOT NULL,
		description TEXT DEFAULT '',
		receipt_path TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS payment_allocations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		payment_id INTEGER NOT NULL,
		advertisement_id INTEGER NOT NULL,
		allocated_amount INTEGER NOT NULL CHECK(allocated_amount > 0),
		FOREIGN KEY(payment_id) REFERENCES payments(id),
		FOREIGN KEY(advertisement_id) REFERENCES advertisements(id)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("creating database tables: %w", err)
	}

	return nil
}
