package ui

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"telegram-ads-manager/internal/database"
	"telegram-ads-manager/internal/models"
)

// App holds the shared state of the application.
type App struct {
	DB      *sql.DB
	FyneApp fyne.App
	Window  fyne.Window

	// Data shared between screens.
	Advertisements []models.Advertisement
	Payments       []models.Payment

	// Advertisement selection.
	SelectedIndex int

	// Payment allocation selection.
	SelectedPaymentID       int64
	SelectedAdvertisementID int64

	// Options used by the allocation screen.
	PaymentOptions       map[string]int64
	AdvertisementOptions map[string]int64

	// Shared page elements.
	PageTitle   *widget.Label
	PageContent *fyne.Container

	// Updated by the allocation screen.
	UpdateAllocationStatus func()

	// Payment allocation screen widgets.
	PaymentSelect       *widget.Select
	AdvertisementSelect *widget.Select
	AllocationStatus    *widget.Label
	AllocationHistory   *widget.Table
	AdvertisementList   *widget.Table
}

// NewApp creates and prepares the application.
func NewApp(db *sql.DB) (*App, error) {
	advertisements, err := database.GetAdvertisements(db)
	if err != nil {
		return nil, err
	}

	payments, err := database.GetPayments(db)
	if err != nil {
		return nil, err
	}

	myApp := app.New()

	window := myApp.NewWindow("Telegram Ads Manager")
	window.Resize(fyne.NewSize(1100, 700))

	newApp := &App{
		DB:                   db,
		FyneApp:              myApp,
		Window:               window,
		Advertisements:       advertisements,
		Payments:             payments,
		SelectedIndex:        -1,
		PaymentOptions:       make(map[string]int64),
		AdvertisementOptions: make(map[string]int64),
	}

	newApp.PageTitle = widget.NewLabelWithStyle(
		"Dashboard",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	newApp.PageContent = container.NewStack(
		newApp.DashboardPage(),
	)

	newApp.BuildNavigation()

	return newApp, nil
}

// Run displays the application window.
func (a *App) Run() {
	a.Window.ShowAndRun()
}
