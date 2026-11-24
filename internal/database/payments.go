package database

import (
	"database/sql"

	"telegram-ads-manager/internal/models"
)

func SavePayment(db *sql.DB, payment *models.Payment) error {
	result, err := db.Exec(`
		INSERT INTO payments (
			amount, payment_date, description, receipt_path
		)
		VALUES (?, ?, ?, ?)
	`,
		payment.Amount,
		payment.PaymentDate,
		payment.Description,
		payment.ReceiptPath,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	payment.ID = id
	return nil
}

func GetPayments(db *sql.DB) ([]models.Payment, error) {
	rows, err := db.Query(`
		SELECT id, amount, payment_date, description, receipt_path
		FROM payments
		ORDER BY payment_date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []models.Payment{}

	for rows.Next() {
		var payment models.Payment

		err := rows.Scan(
			&payment.ID,
			&payment.Amount,
			&payment.PaymentDate,
			&payment.Description,
			&payment.ReceiptPath,
		)
		if err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, rows.Err()
}
