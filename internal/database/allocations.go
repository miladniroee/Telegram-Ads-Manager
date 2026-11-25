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
) (models.PaymentStatus, error) {
	paidAmount, err := GetAdvertisementPaidAmount(db, ad.ID)
	if err != nil {
		return models.Unpaid, err
	}

	if paidAmount == 0 {
		return models.Unpaid, nil
	}

	if paidAmount < ad.Price {
		return models.PartiallyPaid, nil
	}

	return models.Paid, nil
}

// UpdateAllocationAmount changes the amount of an existing allocation.
// It checks that the new amount does not exceed either available balance.
func UpdateAllocationAmount(
	db *sql.DB,
	allocationID int64,
	newAmount int,
) error {
	if newAmount <= 0 {
		return fmt.Errorf("allocation amount must be greater than zero")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var paymentID int64
	var advertisementID int64

	err = tx.QueryRow(`
		SELECT payment_id, advertisement_id
		FROM payment_allocations
		WHERE id = ?
	`, allocationID).Scan(&paymentID, &advertisementID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("allocation not found")
	}
	if err != nil {
		return err
	}

	var paymentAmount int
	err = tx.QueryRow(
		`SELECT amount FROM payments WHERE id = ?`,
		paymentID,
	).Scan(&paymentAmount)
	if err != nil {
		return fmt.Errorf("could not find payment: %w", err)
	}

	var otherPaymentAllocations int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(allocated_amount), 0)
		FROM payment_allocations
		WHERE payment_id = ? AND id != ?
	`, paymentID, allocationID).Scan(&otherPaymentAllocations)
	if err != nil {
		return err
	}

	paymentRemaining := paymentAmount - otherPaymentAllocations
	if newAmount > paymentRemaining {
		return fmt.Errorf(
			"amount exceeds the payment's available balance of %d Toman",
			paymentRemaining,
		)
	}

	var adPrice int
	err = tx.QueryRow(
		`SELECT price FROM advertisements WHERE id = ?`,
		advertisementID,
	).Scan(&adPrice)
	if err != nil {
		return fmt.Errorf("could not find advertisement: %w", err)
	}

	var otherAdvertisementAllocations int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(allocated_amount), 0)
		FROM payment_allocations
		WHERE advertisement_id = ? AND id != ?
	`, advertisementID, allocationID).Scan(&otherAdvertisementAllocations)
	if err != nil {
		return err
	}

	adRemaining := adPrice - otherAdvertisementAllocations
	if newAmount > adRemaining {
		return fmt.Errorf(
			"amount exceeds the advertisement's outstanding balance of %d Toman",
			adRemaining,
		)
	}

	_, err = tx.Exec(`
		UPDATE payment_allocations
		SET allocated_amount = ?
		WHERE id = ?
	`, newAmount, allocationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteAllocation removes an allocation.
// The allocated amount becomes available again for both the payment and advertisement.
func DeleteAllocation(db *sql.DB, allocationID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`DELETE FROM payment_allocations WHERE id = ?`,
		allocationID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("allocation not found")
	}

	return tx.Commit()
}
