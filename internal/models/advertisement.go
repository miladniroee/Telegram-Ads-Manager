package models

type Advertisement struct {
	ID          int64
	Date        string
	Hour        string
	Customer    string
	RemovalDate string
	Price       int
	IsFinished  bool
}
