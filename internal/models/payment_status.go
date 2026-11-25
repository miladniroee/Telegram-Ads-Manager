package models

type PaymentStatus int

const (
	Unpaid        PaymentStatus = 0
	Paid          PaymentStatus = 1
	PartiallyPaid PaymentStatus = 2
)

// Returns the textual label.
func (s PaymentStatus) Label() string {
	switch s {
	case Unpaid:
		return "Unpaid"
	case Paid:
		return "Paid"
	case PartiallyPaid:
		return "Partially paid"
	default:
		return "Unknown"
	}
}

// Returns the corresponding emoji.
func (s PaymentStatus) Emoji() string {
	switch s {
	case Unpaid:
		return "❌"
	case Paid:
		return "✅"
	case PartiallyPaid:
		return "🟡"
	default:
		return "❓"
	}
}

func (s PaymentStatus) IsUnpaid() bool {
	return s == Unpaid
}

func (s PaymentStatus) IsPaid() bool {
	return s == Paid
}

func (s PaymentStatus) IsPartiallyPaid() bool {
	return s == PartiallyPaid
}
