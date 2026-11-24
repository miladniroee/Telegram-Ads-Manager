package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"telegram-ads-manager/internal/database"
	"telegram-ads-manager/internal/models"
)

// AdvertisementsPage creates the advertisement management screen.
func (a *App) AdvertisementsPage() fyne.CanvasObject {
	// Advertisement form fields.
	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("Example: 1405/09/03")

	hourEntry := widget.NewEntry()
	hourEntry.SetPlaceHolder("00–23")

	customerEntry := widget.NewEntry()
	customerEntry.SetPlaceHolder("Customer name")

	removalDateEntry := widget.NewEntry()
	removalDateEntry.SetPlaceHolder("Example: 1405/09/05")

	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Price in Toman")

	finishedCheck := widget.NewCheck("Advertisement removed", nil)

	advertisementList := widget.NewTable(
		func() (int, int) {
			return len(a.Advertisements) + 1, 7
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Advertisement")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			if id.Row == 0 {
				headers := []string{
					"Date", "Hour", "Customer",
					"Removal date", "Price", "Payment status", "Finished",
				}

				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			ad := a.Advertisements[id.Row-1]

			paymentStatus, err := database.GetAdvertisementPaymentStatus(a.DB, ad)
			if err != nil {
				paymentStatus = "Error"
			}

			values := []string{
				ad.Date,
				ad.Hour,
				ad.Customer,
				ad.RemovalDate,
				fmt.Sprintf("%d", ad.Price),
				paymentStatus,
				boolText(ad.IsFinished),
			}

			label.SetText(values[id.Col])
			label.TextStyle = fyne.TextStyle{}
		},
	)

	a.AdvertisementList = advertisementList

	advertisementList.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}

		a.SelectedIndex = id.Row - 1
	}

	advertisementList.SetColumnWidth(0, 100)
	advertisementList.SetColumnWidth(1, 60)
	advertisementList.SetColumnWidth(2, 150)
	advertisementList.SetColumnWidth(3, 110)
	advertisementList.SetColumnWidth(4, 100)
	advertisementList.SetColumnWidth(5, 150)
	advertisementList.SetColumnWidth(6, 90)

	// Add advertisement button.
	addButton := widget.NewButton("Add Advertisement", func() {
		hour, err := strconv.Atoi(hourEntry.Text)
		if err != nil || hour < 0 || hour > 23 {
			dialog.ShowInformation(
				"Validation Error",
				"Please enter a valid hour between 00 and 23.",
				a.Window,
			)
			return
		}

		price, err := strconv.Atoi(priceEntry.Text)
		if err != nil || price < 0 {
			dialog.ShowInformation(
				"Validation Error",
				"Please enter a valid non-negative price.",
				a.Window,
			)
			return
		}

		if dateEntry.Text == "" ||
			customerEntry.Text == "" ||
			removalDateEntry.Text == "" {
			dialog.ShowInformation(
				"Validation Error",
				"Please fill in all required fields.",
				a.Window,
			)
			return
		}

		ad := models.Advertisement{
			Date:        dateEntry.Text,
			Hour:        fmt.Sprintf("%02d", hour),
			Customer:    customerEntry.Text,
			RemovalDate: removalDateEntry.Text,
			Price:       price,
			IsFinished:  finishedCheck.Checked,
		}

		if err := database.SaveAdvertisement(a.DB, &ad); err != nil {
			dialog.ShowError(err, a.Window)
			return
		}

		a.Advertisements = append(a.Advertisements, ad)
		advertisementList.Refresh()

		dateEntry.SetText("")
		hourEntry.SetText("")
		customerEntry.SetText("")
		removalDateEntry.SetText("")
		priceEntry.SetText("")
		finishedCheck.SetChecked(false)
	})

	// Edit selected advertisement button.
	editButton := widget.NewButton("Edit Selected", func() {
		if a.SelectedIndex < 0 ||
			a.SelectedIndex >= len(a.Advertisements) {
			dialog.ShowInformation(
				"No Selection",
				"Please select an advertisement first.",
				a.Window,
			)
			return
		}

		ad := a.Advertisements[a.SelectedIndex]

		editDateEntry := widget.NewEntry()
		editDateEntry.SetText(ad.Date)

		editHourEntry := widget.NewEntry()
		editHourEntry.SetText(ad.Hour)

		editCustomerEntry := widget.NewEntry()
		editCustomerEntry.SetText(ad.Customer)

		editRemovalDateEntry := widget.NewEntry()
		editRemovalDateEntry.SetText(ad.RemovalDate)

		editPriceEntry := widget.NewEntry()
		editPriceEntry.SetText(fmt.Sprintf("%d", ad.Price))

		editFinishedCheck := widget.NewCheck(
			"Advertisement removed",
			nil,
		)
		editFinishedCheck.SetChecked(ad.IsFinished)

		fields := []*widget.FormItem{
			widget.NewFormItem("Advertisement date", editDateEntry),
			widget.NewFormItem("Advertisement hour", editHourEntry),
			widget.NewFormItem("Customer name", editCustomerEntry),
			widget.NewFormItem("Removal date", editRemovalDateEntry),
			widget.NewFormItem("Price (Toman)", editPriceEntry),
			widget.NewFormItem("", editFinishedCheck),
		}

		editDialog := dialog.NewForm(
			"Edit Advertisement",
			"Save Changes",
			"Cancel",
			fields,
			func(confirmed bool) {
				if !confirmed {
					return
				}

				hour, err := strconv.Atoi(editHourEntry.Text)
				if err != nil || hour < 0 || hour > 23 {
					dialog.ShowInformation(
						"Invalid Hour",
						"Enter an hour between 00 and 23.",
						a.Window,
					)
					return
				}

				price, err := strconv.Atoi(editPriceEntry.Text)
				if err != nil || price < 0 {
					dialog.ShowInformation(
						"Invalid Price",
						"Enter a valid non-negative price.",
						a.Window,
					)
					return
				}

				if editDateEntry.Text == "" ||
					editCustomerEntry.Text == "" ||
					editRemovalDateEntry.Text == "" {
					dialog.ShowInformation(
						"Missing Information",
						"Please fill in all required fields.",
						a.Window,
					)
					return
				}

				ad.Date = editDateEntry.Text
				ad.Hour = fmt.Sprintf("%02d", hour)
				ad.Customer = editCustomerEntry.Text
				ad.RemovalDate = editRemovalDateEntry.Text
				ad.Price = price
				ad.IsFinished = editFinishedCheck.Checked

				if err := database.UpdateAdvertisement(a.DB, ad); err != nil {
					dialog.ShowError(err, a.Window)
					return
				}

				a.Advertisements[a.SelectedIndex] = ad
				advertisementList.Refresh()
			},
			a.Window,
		)

		editDialog.Resize(fyne.NewSize(450, 500))
		editDialog.Show()
	})

	// Delete selected advertisement button.
	deleteButton := widget.NewButton("Delete Selected", func() {
		if a.SelectedIndex < 0 ||
			a.SelectedIndex >= len(a.Advertisements) {
			dialog.ShowInformation(
				"No Selection",
				"Please select an advertisement first.",
				a.Window,
			)
			return
		}

		ad := a.Advertisements[a.SelectedIndex]

		dialog.ShowConfirm(
			"Delete Advertisement",
			"Are you sure you want to delete the advertisement for "+ad.Customer+"?",
			func(confirmed bool) {
				if !confirmed {
					return
				}

				if err := database.DeleteAdvertisement(a.DB, ad.ID); err != nil {
					dialog.ShowError(err, a.Window)
					return
				}

				a.Advertisements = append(
					a.Advertisements[:a.SelectedIndex],
					a.Advertisements[a.SelectedIndex+1:]...,
				)

				a.SelectedIndex = -1
				advertisementList.UnselectAll()
				advertisementList.Refresh()
			},
			a.Window,
		)
	})

	// Advertisement form.
	form := widget.NewForm(
		widget.NewFormItem("Advertisement date", dateEntry),
		widget.NewFormItem("Advertisement hour", hourEntry),
		widget.NewFormItem("Customer name", customerEntry),
		widget.NewFormItem("Removal date", removalDateEntry),
		widget.NewFormItem("Price (Toman)", priceEntry),
		widget.NewFormItem("", finishedCheck),
	)

	form.OnSubmit = func() {
		addButton.OnTapped()
	}

	// Complete advertisement page.
	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle(
				"Add Advertisement",
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			),
			form,
			addButton,
			widget.NewSeparator(),
			widget.NewLabelWithStyle(
				"Advertisement List",
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			),
			container.NewHBox(
				editButton,
				deleteButton,
			),
		),
		nil,
		nil,
		nil,
		container.NewScroll(advertisementList),
	)
}

// boolText converts a boolean value to Yes or No.
func boolText(value bool) string {
	if value {
		return "✅"
	}

	return "❌"
}
