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

// PaymentsPage creates the payment management screen.
func (a *App) PaymentsPage() fyne.CanvasObject {
	paymentDateEntry := widget.NewEntry()
	paymentDateEntry.SetPlaceHolder("Example: 1405/09/03")

	paymentAmountEntry := widget.NewEntry()
	paymentAmountEntry.SetPlaceHolder("Amount in Toman")

	paymentDescriptionEntry := widget.NewEntry()
	paymentDescriptionEntry.SetPlaceHolder("Optional description")

	receiptPathEntry := widget.NewEntry()
	receiptPathEntry.SetPlaceHolder("Optional receipt file path")

	paymentList := widget.NewTable(
		func() (int, int) {
			return len(a.Payments) + 1, 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Payment")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			if id.Row == 0 {
				headers := []string{
					"Date", "Amount (Toman)", "Description", "Receipt",
				}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			payment := a.Payments[id.Row-1]

			values := []string{
				payment.PaymentDate,
				fmt.Sprintf("%d", payment.Amount),
				payment.Description,
				payment.ReceiptPath,
			}

			label.SetText(values[id.Col])
			label.TextStyle = fyne.TextStyle{}
		},
	)

	paymentList.SetColumnWidth(0, 110)
	paymentList.SetColumnWidth(1, 150)
	paymentList.SetColumnWidth(2, 200)
	paymentList.SetColumnWidth(3, 200)

	addPaymentButton := widget.NewButton("Register Payment", func() {
		amount, err := strconv.Atoi(paymentAmountEntry.Text)
		if err != nil || amount <= 0 {
			dialog.ShowInformation(
				"Invalid Amount",
				"Please enter a valid payment amount greater than zero.",
				a.Window,
			)
			return
		}

		if paymentDateEntry.Text == "" {
			dialog.ShowInformation(
				"Missing Date",
				"Please enter the payment date.",
				a.Window,
			)
			return
		}

		payment := models.Payment{
			Amount:      amount,
			PaymentDate: paymentDateEntry.Text,
			Description: paymentDescriptionEntry.Text,
			ReceiptPath: receiptPathEntry.Text,
		}

		if err := database.SavePayment(a.DB, &payment); err != nil {
			dialog.ShowError(err, a.Window)
			return
		}

		a.Payments = append(a.Payments, payment)
		paymentList.Refresh()

		paymentDateEntry.SetText("")
		paymentAmountEntry.SetText("")
		paymentDescriptionEntry.SetText("")
		receiptPathEntry.SetText("")

		dialog.ShowInformation(
			"Payment Saved",
			"Payment was successfully registered.",
			a.Window,
		)
	})

	paymentForm := container.NewVBox(
		widget.NewLabelWithStyle(
			"Register New Payment",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewForm(
			widget.NewFormItem("Payment date", paymentDateEntry),
			widget.NewFormItem("Amount (Toman)", paymentAmountEntry),
			widget.NewFormItem("Description", paymentDescriptionEntry),
			widget.NewFormItem("Receipt path", receiptPathEntry),
		),
		addPaymentButton,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(
			"Payment History",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
	)

	return container.NewBorder(
		paymentForm,
		nil,
		nil,
		nil,
		container.NewScroll(paymentList),
	)
}
