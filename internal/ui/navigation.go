package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowPage changes the content displayed in the main area.
func (a *App) ShowPage(title string, content fyne.CanvasObject) {
	a.PageTitle.SetText(title)
	a.PageContent.Objects = []fyne.CanvasObject{content}
	a.PageContent.Refresh()
}

// BuildNavigation creates the main application layout and sidebar.
func (a *App) BuildNavigation() {
	sidebar := container.NewVBox(
		widget.NewLabelWithStyle(
			"TELEGRAM ADS",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),

		widget.NewButton("Dashboard", func() {
			a.ShowPage("Dashboard", a.DashboardPage())
		}),

		widget.NewButton("Advertisements", func() {
			a.ShowPage("Advertisements", a.AdvertisementsPage())
		}),

		widget.NewButton("Payments", func() {
			a.ShowPage("Payments", a.PaymentsPage())
		}),

		widget.NewButton("Payment Allocations", func() {
			a.RefreshAllocationOptions()
			a.ShowPage("Payment Allocations", a.AllocationsPage())
		}),

		widget.NewButton("Settings", func() {
			a.ShowPage("Settings", a.SettingsPage())
		}),
	)

	mainArea := container.NewBorder(
		container.NewPadded(a.PageTitle),
		nil,
		nil,
		nil,
		container.NewPadded(a.PageContent),
	)

	layout := container.NewBorder(
		nil,
		nil,
		container.NewPadded(sidebar),
		nil,
		mainArea,
	)

	a.Window.SetContent(layout)
}
