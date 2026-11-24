package models

type Payment struct {
	ID          int64
	Amount      int
	PaymentDate string
	Description string
	ReceiptPath string
}
