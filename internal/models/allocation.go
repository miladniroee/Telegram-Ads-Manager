package models

type Allocation struct {
	ID              int64
	PaymentID       int64
	AdvertisementID int64
	CustomerName    string
	Amount          int
	PaymentDate     string
}
