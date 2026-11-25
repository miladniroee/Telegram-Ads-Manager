package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"telegram-ads-manager/internal/database"
)

// SettingsPage creates the settings screen.
func (a *App) SettingsPage() fyne.CanvasObject {
	// Database backup
	backupPathEntry := widget.NewEntry()
	backupPathEntry.SetText("telegram_ads_backup.db")
	backupPathEntry.SetPlaceHolder("Enter backup file path")

	backupStatusLabel := widget.NewLabel("")

	createBackupButton := widget.NewButton(
		"Create Database Backup",
		func() {
			err := database.CreateBackup(a.DB, backupPathEntry.Text)
			if err != nil {
				dialog.ShowError(err, a.Window)
				return
			}

			backupStatusLabel.SetText("Backup created successfully.")
			dialog.ShowInformation(
				"Backup Created",
				"Your database backup was created successfully.",
				a.Window,
			)
		},
	)

	// Database restore
	restorePathEntry := widget.NewEntry()
	restorePathEntry.SetPlaceHolder("Enter the backup file path")

	restoreStatusLabel := widget.NewLabel("")

	restoreBackupButton := widget.NewButton(
		"Restore Database Backup",
		func() {
			if restorePathEntry.Text == "" {
				dialog.ShowInformation(
					"Missing Backup Path",
					"Please enter the path to a backup file.",
					a.Window,
				)
				return
			}

			dialog.ShowConfirm(
				"Restore Database",
				"This will replace all current advertisements, payments, and allocations with the data from the backup. Are you sure?",
				func(confirmed bool) {
					if !confirmed {
						return
					}

					err := database.RestoreBackup(a.DB, restorePathEntry.Text)
					if err != nil {
						dialog.ShowError(err, a.Window)
						return
					}

					advertisements, err := database.GetAdvertisements(a.DB)
					if err != nil {
						dialog.ShowError(err, a.Window)
						return
					}

					payments, err := database.GetPayments(a.DB)
					if err != nil {
						dialog.ShowError(err, a.Window)
						return
					}

					a.Advertisements = advertisements
					a.Payments = payments

					a.SelectedPaymentID = 0
					a.SelectedAdvertisementID = 0

					a.RefreshAllocationOptions()

					if a.AdvertisementList != nil {
						a.AdvertisementList.Refresh()
					}

					if a.AllocationHistory != nil {
						a.AllocationHistory.Refresh()
					}

					restoreStatusLabel.SetText("Database restored successfully.")

					dialog.ShowInformation(
						"Restore Complete",
						"Your database has been restored successfully.",
						a.Window,
					)
				},
				a.Window,
			)
		},
	)

	// CSV export
	advertisementExportEntry := widget.NewEntry()
	advertisementExportEntry.SetText("advertisements.csv")

	paymentExportEntry := widget.NewEntry()
	paymentExportEntry.SetText("payments.csv")

	allocationExportEntry := widget.NewEntry()
	allocationExportEntry.SetText("allocations.csv")

	exportButton := func(
		buttonLabel string,
		entry *widget.Entry,
		export func(string) error,
	) *widget.Button {
		return widget.NewButton(buttonLabel, func() {
			err := export(entry.Text)
			if err != nil {
				dialog.ShowError(err, a.Window)
				return
			}

			dialog.ShowInformation(
				"Export Complete",
				"CSV file created successfully:\n"+entry.Text,
				a.Window,
			)
		})
	}

	exportAdvertisementsButton := exportButton(
		"Export Advertisements",
		advertisementExportEntry,
		func(path string) error {
			return database.ExportAdvertisementsCSV(a.DB, path)
		},
	)

	exportPaymentsButton := exportButton(
		"Export Payments",
		paymentExportEntry,
		func(path string) error {
			return database.ExportPaymentsCSV(a.DB, path)
		},
	)

	exportAllocationsButton := exportButton(
		"Export Allocations",
		allocationExportEntry,
		func(path string) error {
			return database.ExportAllocationsCSV(a.DB, path)
		},
	)

	return container.NewScroll(container.NewVBox(
		widget.NewLabelWithStyle(
			"Settings",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),

		widget.NewLabelWithStyle(
			"Database Backup",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewLabel(
			"Create a backup of your advertisements, payments, and payment allocations.",
		),
		backupPathEntry,
		createBackupButton,
		backupStatusLabel,

		widget.NewSeparator(),

		widget.NewLabelWithStyle(
			"Restore Database Backup",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewLabel(
			"Restore your database from an existing backup. This replaces the current records.",
		),
		restorePathEntry,
		restoreBackupButton,
		restoreStatusLabel,

		widget.NewSeparator(),

		widget.NewLabelWithStyle(
			"Export to CSV",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewLabel(
			"Export your records to CSV files that can be opened in LibreOffice Calc or Excel.",
		),

		widget.NewLabel("Advertisements"),
		advertisementExportEntry,
		exportAdvertisementsButton,

		widget.NewLabel("Payments"),
		paymentExportEntry,
		exportPaymentsButton,

		widget.NewLabel("Payment Allocations"),
		allocationExportEntry,
		exportAllocationsButton,
	))
}
