package database

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ExportAdvertisementsCSV exports all advertisements to a CSV file.
func ExportAdvertisementsCSV(db *sql.DB, destination string) error {
	advertisements, err := GetAdvertisements(db)
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(advertisements))

	for _, ad := range advertisements {
		paymentStatus, err := GetAdvertisementPaymentStatus(db, ad)
		if err != nil {
			return err
		}

		rows = append(rows, []string{
			strconv.FormatInt(ad.ID, 10),
			ad.Date,
			ad.Hour,
			safeCSVText(ad.Customer),
			ad.RemovalDate,
			strconv.Itoa(ad.Price),
			strconv.FormatBool(ad.IsFinished),
			paymentStatus.Label(),
		})
	}

	return writeCSV(destination, []string{
		"ID",
		"Ad Date",
		"Hour",
		"Customer",
		"Removal Date",
		"Price (Toman)",
		"Finished",
		"Payment Status",
	}, rows)
}

// ExportPaymentsCSV exports all payments to a CSV file.
func ExportPaymentsCSV(db *sql.DB, destination string) error {
	payments, err := GetPayments(db)
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(payments))

	for _, payment := range payments {
		remaining, err := GetPaymentRemaining(db, payment.ID)
		if err != nil {
			return err
		}

		rows = append(rows, []string{
			strconv.FormatInt(payment.ID, 10),
			payment.PaymentDate,
			strconv.Itoa(payment.Amount),
			strconv.Itoa(payment.Amount - remaining),
			strconv.Itoa(remaining),
			safeCSVText(payment.Description),
			safeCSVText(payment.ReceiptPath),
		})
	}

	return writeCSV(destination, []string{
		"ID",
		"Payment Date",
		"Amount (Toman)",
		"Allocated (Toman)",
		"Remaining (Toman)",
		"Description",
		"Receipt Path",
	}, rows)
}

// ExportAllocationsCSV exports all payment allocations to a CSV file.
func ExportAllocationsCSV(db *sql.DB, destination string) error {
	allocations, err := GetPaymentAllocations(db)
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(allocations))

	for _, allocation := range allocations {
		rows = append(rows, []string{
			strconv.FormatInt(allocation.ID, 10),
			strconv.FormatInt(allocation.PaymentID, 10),
			strconv.FormatInt(allocation.AdvertisementID, 10),
			safeCSVText(allocation.CustomerName),
			strconv.Itoa(allocation.Amount),
			allocation.PaymentDate,
		})
	}

	return writeCSV(destination, []string{
		"Allocation ID",
		"Payment ID",
		"Advertisement ID",
		"Customer",
		"Allocated Amount (Toman)",
		"Payment Date",
	}, rows)
}

// writeCSV creates a CSV file without overwriting an existing file.
func writeCSV(destination string, headers []string, rows [][]string) error {
	destination = strings.TrimSpace(destination)
	if destination == "" {
		return fmt.Errorf("please enter an export file path")
	}

	absolutePath, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("invalid export path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0755); err != nil {
		return fmt.Errorf("could not create export directory: %w", err)
	}

	file, err := os.OpenFile(
		absolutePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("a file already exists at this path")
		}
		return err
	}

	writer := csv.NewWriter(file)

	if err := writer.Write(headers); err != nil {
		file.Close()
		os.Remove(absolutePath)
		return err
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			file.Close()
			os.Remove(absolutePath)
			return err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		file.Close()
		os.Remove(absolutePath)
		return err
	}

	if err := file.Close(); err != nil {
		os.Remove(absolutePath)
		return err
	}

	return nil
}

// safeCSVText helps prevent spreadsheet formula injection in text fields.
func safeCSVText(value string) string {
	if value == "" {
		return value
	}

	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}
