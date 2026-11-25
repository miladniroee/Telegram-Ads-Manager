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

	statusLabel := widget.NewLabel("")

	createBackupButton := widget.NewButton(
		"Create Database Backup",
		func() {
			err := database.CreateBackup(a.DB, backupPathEntry.Text)
			if err != nil {
				dialog.ShowError(err, a.Window)
				return
			}

			statusLabel.SetText("Backup created successfully.")

			dialog.ShowInformation(
				"Backup Created",
				"Your database backup was created successfully.",
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
		statusLabel,
	)
}
