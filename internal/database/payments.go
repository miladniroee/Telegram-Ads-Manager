package database

import (
	"database/sql"
	"fmt"

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

// GetPaymentAllocatedAmount returns the total amount allocated from a payment.
func GetPaymentAllocatedAmount(db *sql.DB, paymentID int64) (int, error) {
	var total int

	err := db.QueryRow(`
		SELECT COALESCE(SUM(allocated_amount), 0)
		FROM payment_allocations
		WHERE payment_id = ?
	`, paymentID).Scan(&total)

	return total, err
}

// UpdatePayment updates a payment without violating existing allocations.
func UpdatePayment(db *sql.DB, payment *models.Payment) error {
	if payment.Amount <= 0 {
		return fmt.Errorf("payment amount must be greater than zero")
	}

	allocatedAmount, err := GetPaymentAllocatedAmount(db, payment.ID)
	if err != nil {
		return err
	}

	if payment.Amount < allocatedAmount {
		return fmt.Errorf(
			"payment amount cannot be less than the allocated amount of %d Toman",
			allocatedAmount,
		)
	}

	_, err = db.Exec(`
		UPDATE payments
		SET amount = ?, payment_date = ?, description = ?, receipt_path = ?
		WHERE id = ?
	`,
		payment.Amount,
		payment.PaymentDate,
		payment.Description,
		payment.ReceiptPath,
		payment.ID,
	)

	return err
}

// DeletePayment only deletes payments that have no allocations.
func DeletePayment(db *sql.DB, paymentID int64) error {
	allocatedAmount, err := GetPaymentAllocatedAmount(db, paymentID)
	if err != nil {
		return err
	}

	if allocatedAmount > 0 {
		return fmt.Errorf(
			"this payment has %d Toman allocated to advertisements; handle its allocations before deleting it",
			allocatedAmount,
		)
	}

	result, err := db.Exec(
		`DELETE FROM payments WHERE id = ?`,
		paymentID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}
