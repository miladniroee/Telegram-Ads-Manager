package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SettingsPage creates the settings screen.
func (a *App) SettingsPage() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle(
			"Settings",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),
		widget.NewLabel("Application settings will be available here."),
	)
}
