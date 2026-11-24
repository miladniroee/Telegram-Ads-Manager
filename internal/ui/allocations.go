package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"telegram-ads-manager/internal/database"
)

// AllocationsPage creates the payment allocation screen.
func (a *App) AllocationsPage() fyne.CanvasObject {
	a.PaymentSelect = widget.NewSelect(
		[]string{},
		func(value string) {
			a.SelectedPaymentID = a.PaymentOptions[value]
			a.updateAllocationStatus()
		},
	)

	a.AdvertisementSelect = widget.NewSelect(
		[]string{},
		func(value string) {
			a.SelectedAdvertisementID = a.AdvertisementOptions[value]
			a.updateAllocationStatus()
		},
	)

	allocationAmountEntry := widget.NewEntry()
	allocationAmountEntry.SetPlaceHolder("Amount in Toman")

	a.AllocationStatus = widget.NewLabel(
		"Select a payment and an advertisement.",
	)

	a.AllocationHistory = widget.NewTable(
		func() (int, int) {
			allocations, err := database.GetPaymentAllocations(a.DB)
			if err != nil {
				return 0, 5
			}
			return len(allocations) + 1, 5
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Allocation")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			if id.Row == 0 {
				headers := []string{
					"Payment ID",
					"Ad ID",
					"Customer",
					"Amount (Toman)",
					"Payment Date",
				}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			allocations, err := database.GetPaymentAllocations(a.DB)
			if err != nil || id.Row-1 >= len(allocations) {
				label.SetText("")
				return
			}

			allocation := allocations[id.Row-1]

			values := []string{
				fmt.Sprintf("%d", allocation.PaymentID),
				fmt.Sprintf("%d", allocation.AdvertisementID),
				allocation.CustomerName,
				fmt.Sprintf("%d", allocation.Amount),
				allocation.PaymentDate,
			}

			label.SetText(values[id.Col])
			label.TextStyle = fyne.TextStyle{}
		},
	)

	addAllocationButton := widget.NewButton(
		"Allocate Payment",
		func() {
			if a.SelectedPaymentID == 0 ||
				a.SelectedAdvertisementID == 0 {
				dialog.ShowInformation(
					"Missing Selection",
					"Please select both a payment and an advertisement.",
					a.Window,
				)
				return
			}

			amount, err := strconv.Atoi(allocationAmountEntry.Text)
			if err != nil || amount <= 0 {
				dialog.ShowInformation(
					"Invalid Amount",
					"Enter a valid amount greater than zero.",
					a.Window,
				)
				return
			}

			err = database.AllocatePayment(
				a.DB,
				a.SelectedPaymentID,
				a.SelectedAdvertisementID,
				amount,
			)
			if err != nil {
				dialog.ShowError(err, a.Window)
				a.updateAllocationStatus()
				return
			}

			a.AllocationHistory.Refresh()
			allocationAmountEntry.SetText("")
			a.updateAllocationStatus()

			// Update the selected advertisement's payment status.
			for i := range a.Advertisements {
				if a.Advertisements[i].ID == a.SelectedAdvertisementID {
					balance, err := database.GetAdvertisementBalance(
						a.DB,
						a.SelectedAdvertisementID,
					)
					if err == nil {
						a.Advertisements[i].IsPaid = balance == 0
					}
					break
				}
			}

			dialog.ShowInformation(
				"Allocation Saved",
				"Payment was successfully allocated.",
				a.Window,
			)
		},
	)

	// Load the available payments and advertisements.
	a.RefreshAllocationOptions()

	return container.NewVBox(
		widget.NewLabelWithStyle(
			"Allocate a Payment to an Advertisement",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		a.PaymentSelect,
		a.AdvertisementSelect,
		allocationAmountEntry,
		a.AllocationStatus,
		addAllocationButton,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(
			"Allocation History",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		container.NewScroll(a.AllocationHistory),
	)
}

// RefreshAllocationOptions refreshes the available payments and advertisements.
func (a *App) RefreshAllocationOptions() {
	if a.PaymentSelect == nil || a.AdvertisementSelect == nil {
		return
	}

	a.PaymentOptions = make(map[string]int64)
	a.AdvertisementOptions = make(map[string]int64)

	paymentLabels := []string{}

	for _, payment := range a.Payments {
		label := fmt.Sprintf(
			"Payment #%d — %d Toman (%s)",
			payment.ID,
			payment.Amount,
			payment.PaymentDate,
		)

		a.PaymentOptions[label] = payment.ID
		paymentLabels = append(paymentLabels, label)
	}

	advertisementLabels := []string{}

	for _, ad := range a.Advertisements {
		label := fmt.Sprintf(
			"Ad #%d — %s — %s",
			ad.ID,
			ad.Customer,
			ad.Date,
		)

		a.AdvertisementOptions[label] = ad.ID
		advertisementLabels = append(advertisementLabels, label)
	}

	a.PaymentSelect.SetOptions(paymentLabels)
	a.AdvertisementSelect.SetOptions(advertisementLabels)

	a.SelectedPaymentID = 0
	a.SelectedAdvertisementID = 0

	a.PaymentSelect.ClearSelected()
	a.AdvertisementSelect.ClearSelected()

	a.updateAllocationStatus()

	if a.AllocationHistory != nil {
		a.AllocationHistory.Refresh()
	}
}

// updateAllocationStatus displays the remaining payment and advertisement balances.
func (a *App) updateAllocationStatus() {
	if a.AllocationStatus == nil {
		return
	}

	if a.SelectedPaymentID == 0 || a.SelectedAdvertisementID == 0 {
		a.AllocationStatus.SetText(
			"Select a payment and an advertisement.",
		)
		return
	}

	paymentRemaining, err := database.GetPaymentRemaining(
		a.DB,
		a.SelectedPaymentID,
	)
	if err != nil {
		a.AllocationStatus.SetText("Could not read payment balance.")
		return
	}

	adBalance, err := database.GetAdvertisementBalance(
		a.DB,
		a.SelectedAdvertisementID,
	)
	if err != nil {
		a.AllocationStatus.SetText("Could not read advertisement balance.")
		return
	}

	a.AllocationStatus.SetText(fmt.Sprintf(
		"Remaining payment: %d Toman\nOutstanding advertisement balance: %d Toman",
		paymentRemaining,
		adBalance,
	))
}
