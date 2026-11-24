package database

import (
	"database/sql"
	"fmt"

	"telegram-ads-manager/internal/models"
)

func SaveAdvertisement(db *sql.DB, ad *models.Advertisement) error {
	result, err := db.Exec(`
		INSERT INTO advertisements (
			ad_date, ad_time, customer_name,
			removal_date, price, is_finished
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		ad.Date,
		ad.Hour,
		ad.Customer,
		ad.RemovalDate,
		ad.Price,
		boolToInt(ad.IsFinished),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	ad.ID = id
	return nil
}

func GetAdvertisements(db *sql.DB) ([]models.Advertisement, error) {
	rows, err := db.Query(`
		SELECT id, ad_date, ad_time, customer_name,
		       removal_date, price, is_finished
		FROM advertisements
		ORDER BY ad_date DESC, ad_time DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	advertisements := []models.Advertisement{}

	for rows.Next() {
		var ad models.Advertisement
		var hour, isFinished int

		err := rows.Scan(
			&ad.ID,
			&ad.Date,
			&hour,
			&ad.Customer,
			&ad.RemovalDate,
			&ad.Price,
			&isFinished,
		)
		if err != nil {
			return nil, err
		}

		ad.Hour = fmt.Sprintf("%02d", hour)
		ad.IsFinished = isFinished == 1

		advertisements = append(advertisements, ad)
	}

	return advertisements, rows.Err()
}

func UpdateAdvertisement(db *sql.DB, ad models.Advertisement) error {
	_, err := db.Exec(`
		UPDATE advertisements
		SET ad_date = ?, ad_time = ?, customer_name = ?,
		    removal_date = ?, price = ?, is_finished = ?
		WHERE id = ?
	`,
		ad.Date,
		ad.Hour,
		ad.Customer,
		ad.RemovalDate,
		ad.Price,
		boolToInt(ad.IsFinished),
		ad.ID,
	)

	return err
}

func DeleteAdvertisement(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		DELETE FROM advertisements WHERE id = ?
	`, id)

	return err
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
