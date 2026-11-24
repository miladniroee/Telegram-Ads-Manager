package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DashboardPage creates the dashboard screen.
func (a *App) DashboardPage() *fyne.Container {
	return container.NewVBox(
		widget.NewLabelWithStyle(
			"Welcome to your Telegram Ads Manager!",
			0,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),
		widget.NewLabel("Your application dashboard."),
	)
}
