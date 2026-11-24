package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"telegram-ads-manager/internal/database"
)

// DashboardPage creates the dashboard screen.
func (a *App) DashboardPage() *fyne.Container {
	stats, err := database.GetDashboardStats(a.DB)
	if err != nil {
		return container.NewVBox(
			widget.NewLabelWithStyle(
				"Dashboard",
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			),
			widget.NewLabel("Could not load dashboard statistics: "+err.Error()),
		)
	}

	totalAdsCard := widget.NewCard(
		"Total Advertisements",
		"",
		widget.NewLabel(fmt.Sprintf("%d", stats.TotalAdvertisements)),
	)

	activeAdsCard := widget.NewCard(
		"Active Advertisements",
		"",
		widget.NewLabel(fmt.Sprintf("%d", stats.ActiveAdvertisements)),
	)

	completedAdsCard := widget.NewCard(
		"Completed Advertisements",
		"",
		widget.NewLabel(fmt.Sprintf("%d", stats.CompletedAdvertisements)),
	)

	totalValueCard := widget.NewCard(
		"Total Advertisement Value",
		"",
		widget.NewLabel(fmt.Sprintf("%d Toman", stats.TotalAdValue)),
	)

	receivedCard := widget.NewCard(
		"Received Payments",
		"",
		widget.NewLabel(fmt.Sprintf("%d Toman", stats.AllocatedPayments)),
	)

	outstandingCard := widget.NewCard(
		"Outstanding Balance",
		"",
		widget.NewLabel(fmt.Sprintf("%d Toman", stats.OutstandingBalance)),
	)

	statsGrid := container.NewGridWithColumns(
		3,
		totalAdsCard,
		activeAdsCard,
		completedAdsCard,
		totalValueCard,
		receivedCard,
		outstandingCard,
	)

	return container.NewVBox(
		widget.NewLabelWithStyle(
			"Dashboard Overview",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewLabel("Here is an overview of your advertising business."),
		widget.NewSeparator(),
		statsGrid,
	)
}
