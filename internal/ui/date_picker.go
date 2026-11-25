package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"telegram-ads-manager/internal/dates"
)

// datePickerField adds a button beside an existing Entry.
// The Entry continues to hold the selected date as YYYY/MM/DD.
func datePickerField(parent fyne.Window, entry *widget.Entry) *fyne.Container {
	selectedValue := entry.Text

	// Keep the field visually enabled, but prevent manual editing.
	entry.OnChanged = func(value string) {
		if value != selectedValue {
			entry.SetText(selectedValue)
		}
	}

	setDate := func(value string) {
		selectedValue = value
		entry.SetText(value)
	}

	todayButton := widget.NewButton("Today", func() {
		today, err := dates.GregorianToJalali(
			time.Now().Format("2006-01-02"),
		)
		if err != nil {
			return
		}

		setDate(today)
	})

	chooseButton := widget.NewButton("Choose date", func() {
		openDatePicker(parent, entry, setDate)
	})

	buttons := container.NewHBox(todayButton, chooseButton)

	return container.NewBorder(nil, nil, nil, buttons, entry)
}

func openDatePicker(
	parent fyne.Window,
	entry *widget.Entry,
	setDate func(string),
) {
	initialDate := time.Now()

	// Accept an existing Jalali or Gregorian date.
	if gregorian, err := dates.JalaliToGregorian(entry.Text); err == nil {
		if parsed, err := time.Parse("2006-01-02", gregorian); err == nil {
			initialDate = parsed
		}
	} else if parsed, err := time.Parse("2006-01-02", entry.Text); err == nil {
		initialDate = parsed
	}

	initialJalali, err := dates.GregorianToJalali(initialDate.Format("2006-01-02"))
	if err != nil {
		return
	}

	var pickerDialog *dialog.CustomDialog

	modeSelect := widget.NewSelect(
		[]string{"Jalali", "Gregorian"},
		nil,
	)
	modeSelect.SetSelected("Jalali")

	calendarContent := container.NewVBox()

	// Called when the user selects a date.
	selectedDateLabel := widget.NewLabel("Selected date: " + entry.Text)

	if entry.Text == "" {
		selectedDateLabel.SetText("No date selected")
	}

	selectDate := func(value string) {
		setDate(value)
		selectedDateLabel.SetText("Selected date: " + value)
		pickerDialog.Hide()
	}

	// Build the Jalali calendar.
	var showJalaliCalendar func(year, month int)

	showJalaliCalendar = func(year, month int) {
		calendarContent.Objects = nil

		monthNames := []string{
			"Farvardin", "Ordibehesht", "Khordad",
			"Tir", "Mordad", "Shahrivar",
			"Mehr", "Aban", "Azar",
			"Dey", "Bahman", "Esfand",
		}

		title := widget.NewLabel(
			fmt.Sprintf("%s %d", monthNames[month-1], year),
		)

		previous := widget.NewButton("<", func() {
			month--
			if month < 1 {
				month = 12
				year--
			}
			showJalaliCalendar(year, month)
		})

		next := widget.NewButton(">", func() {
			month++
			if month > 12 {
				month = 1
				year++
			}
			showJalaliCalendar(year, month)
		})

		calendarContent.Add(
			container.NewBorder(nil, nil, previous, next, title),
		)

		weekdays := []string{
			"ش", "ی", "د", "س", "چ", "پ", "ج",
		}

		dayCells := []fyne.CanvasObject{}
		for _, weekday := range weekdays {
			dayCells = append(dayCells, widget.NewLabel(weekday))
		}

		firstDate, err := dates.JalaliToGregorian(
			fmt.Sprintf("%04d/%02d/01", year, month),
		)
		if err != nil {
			return
		}

		firstGregorian, err := time.Parse("2006-01-02", firstDate)
		if err != nil {
			return
		}

		// Go: Sunday=0, Saturday=6.
		// Jalali calendar: Saturday=0.
		startOffset := (int(firstGregorian.Weekday()) + 1) % 7

		nextYear, nextMonth := year, month+1
		if nextMonth > 12 {
			nextYear++
			nextMonth = 1
		}

		nextMonthDate, err := dates.JalaliToGregorian(
			fmt.Sprintf("%04d/%02d/01", nextYear, nextMonth),
		)
		if err != nil {
			return
		}

		nextGregorian, err := time.Parse("2006-01-02", nextMonthDate)
		if err != nil {
			return
		}

		daysInMonth := int(nextGregorian.Sub(firstGregorian).Hours() / 24)

		for i := 0; i < startOffset; i++ {
			dayCells = append(dayCells, widget.NewLabel(""))
		}

		for day := 1; day <= daysInMonth; day++ {
			selectedDay := day
			dateValue := fmt.Sprintf(
				"%04d/%02d/%02d",
				year,
				month,
				selectedDay,
			)

			dayCells = append(dayCells, widget.NewButton(
				fmt.Sprintf("%d", selectedDay),
				func() {
					selectDate(dateValue)
				},
			))
		}

		for len(dayCells)%7 != 0 {
			dayCells = append(dayCells, widget.NewLabel(""))
		}

		calendarContent.Add(container.NewGridWithColumns(7, dayCells...))
		calendarContent.Refresh()
	}

	showGregorianCalendar := func() {
		calendarContent.Objects = nil

		calendar := widget.NewCalendar(initialDate, func(selected time.Time) {
			jalaliDate, err := dates.GregorianToJalali(
				selected.Format("2006-01-02"),
			)
			if err == nil {
				selectDate(jalaliDate)
			}
		})

		calendarContent.Add(calendar)
		calendarContent.Refresh()
	}

	modeSelect.OnChanged = func(mode string) {
		if mode == "Gregorian" {
			showGregorianCalendar()
		} else {
			year := initialDate.Year()
			month := int(initialDate.Month())
			_ = year
			_ = month

			// Recalculate the Jalali year and month.
			parts := initialJalali
			var jalaliYear, jalaliMonth, jalaliDay int
			_, _ = fmt.Sscanf(parts, "%d/%d/%d", &jalaliYear, &jalaliMonth, &jalaliDay)

			showJalaliCalendar(jalaliYear, jalaliMonth)
		}
	}

	// Start with the Jalali calendar.
	var jalaliYear, jalaliMonth, jalaliDay int
	_, _ = fmt.Sscanf(initialJalali, "%d/%d/%d", &jalaliYear, &jalaliMonth, &jalaliDay)
	showJalaliCalendar(jalaliYear, jalaliMonth)

	content := container.NewVBox(
		modeSelect,
		selectedDateLabel,
		widget.NewSeparator(),
		calendarContent,
	)

	pickerDialog = dialog.NewCustom("Select date", "Close", content, parent)
	pickerDialog.Resize(fyne.NewSize(360, 420))
	pickerDialog.Show()
}
