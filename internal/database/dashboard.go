package database

import (
	"database/sql"

	"telegram-ads-manager/internal/models"
)

func GetDashboardStats(db *sql.DB) (models.DashboardStats, error) {
	var stats models.DashboardStats

	err := db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN is_finished = 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_finished = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(price), 0),
			(
				SELECT COALESCE(SUM(allocated_amount), 0)
				FROM payment_allocations
			)
		FROM advertisements
	`).Scan(
		&stats.TotalAdvertisements,
		&stats.ActiveAdvertisements,
		&stats.CompletedAdvertisements,
		&stats.TotalAdValue,
		&stats.AllocatedPayments,
	)

	if err != nil {
		return stats, err
	}

	stats.OutstandingBalance = stats.TotalAdValue - stats.AllocatedPayments

	return stats, nil
}
