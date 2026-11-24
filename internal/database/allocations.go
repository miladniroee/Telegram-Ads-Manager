package database

import (
	"database/sql"
	"fmt"

	"telegram-ads-manager/internal/models"
)

func AllocatePayment(
	db *sql.DB,
	paymentID int64,
	advertisementID int64,
	amount int,
) error {
	if amount <= 0 {
		return fmt.Errorf("allocation amount must be greater than zero")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var paymentAmount int
	err = tx.QueryRow(
		`SELECT amount FROM payments WHERE id = ?`,
		paymentID,
	).Scan(&paymentAmount)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	var allocatedPayment int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(allocated_amount), 0)
		FROM payment_allocations
		WHERE payment_id = ?
	`, paymentID).Scan(&allocatedPayment)
	if err != nil {
		return err
	}

	if amount > paymentAmount-allocatedPayment {
		return fmt.Errorf(
			"allocation exceeds the remaining payment amount of %d Toman",
			paymentAmount-allocatedPayment,
		)
	}

	var adPrice int
	err = tx.QueryRow(
		`SELECT price FROM advertisements WHERE id = ?`,
		advertisementID,
	).Scan(&adPrice)
	if err != nil {
		return fmt.Errorf("advertisement not found: %w", err)
	}

	var allocatedAdvertisement int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(allocated_amount), 0)
		FROM payment_allocations
		WHERE advertisement_id = ?
	`, advertisementID).Scan(&allocatedAdvertisement)
	if err != nil {
		return err
	}

	if amount > adPrice-allocatedAdvertisement {
		return fmt.Errorf(
			"allocation exceeds the advertisement's outstanding balance of %d Toman",
			adPrice-allocatedAdvertisement,
		)
	}

	_, err = tx.Exec(`
		INSERT INTO payment_allocations (
			payment_id, advertisement_id, allocated_amount
		)
		VALUES (?, ?, ?)
	`, paymentID, advertisementID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetPaymentRemaining(db *sql.DB, paymentID int64) (int, error) {
	var remaining int

	err := db.QueryRow(`
		SELECT p.amount - COALESCE(SUM(pa.allocated_amount), 0)
		FROM payments p
		LEFT JOIN payment_allocations pa ON pa.payment_id = p.id
		WHERE p.id = ?
		GROUP BY p.id
	`, paymentID).Scan(&remaining)

	return remaining, err
}

func GetAdvertisementBalance(db *sql.DB, advertisementID int64) (int, error) {
	var balance int

	err := db.QueryRow(`
		SELECT a.price - COALESCE(SUM(pa.allocated_amount), 0)
		FROM advertisements a
		LEFT JOIN payment_allocations pa ON pa.advertisement_id = a.id
		WHERE a.id = ?
		GROUP BY a.id
	`, advertisementID).Scan(&balance)

	return balance, err
}

func GetPaymentAllocations(db *sql.DB) ([]models.Allocation, error) {
	rows, err := db.Query(`
		SELECT
			pa.id,
			pa.payment_id,
			pa.advertisement_id,
			a.customer_name,
			pa.allocated_amount,
			p.payment_date
		FROM payment_allocations pa
		JOIN advertisements a ON a.id = pa.advertisement_id
		JOIN payments p ON p.id = pa.payment_id
		ORDER BY pa.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allocations := []models.Allocation{}

	for rows.Next() {
		var allocation models.Allocation

		err := rows.Scan(
			&allocation.ID,
			&allocation.PaymentID,
			&allocation.AdvertisementID,
			&allocation.CustomerName,
			&allocation.Amount,
			&allocation.PaymentDate,
		)
		if err != nil {
			return nil, err
		}

		allocations = append(allocations, allocation)
	}

	return allocations, rows.Err()
}

func GetAdvertisementPaidAmount(db *sql.DB, advertisementID int64) (int, error) {
	var total int

	err := db.QueryRow(`
		SELECT COALESCE(SUM(allocated_amount), 0)
		FROM payment_allocations
		WHERE advertisement_id = ?
	`, advertisementID).Scan(&total)

	return total, err
}

func GetAdvertisementPaymentStatus(
	db *sql.DB,
	ad models.Advertisement,
) (string, error) {
	paidAmount, err := GetAdvertisementPaidAmount(db, ad.ID)
	if err != nil {
		return "", err
	}

	if paidAmount == 0 {
		return "Unpaid", nil
	}

	if paidAmount < ad.Price {
		return "🟡", nil
	}

	return "Paid", nil
}
