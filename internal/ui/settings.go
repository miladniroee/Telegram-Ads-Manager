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

					err := database.RestoreBackup(
						a.DB,
						restorePathEntry.Text,
					)
					if err != nil {
						dialog.ShowError(err, a.Window)
						return
					}

					// Reload the in-memory data after restoring.
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

	return container.NewVBox(
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
	)
}
