package main

import (
	"telegram-ads-manager/internal/database"
	"telegram-ads-manager/internal/ui"
)

func main() {
	// Open the database.
	db, err := database.OpenDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create the application interface.
	application, err := ui.NewApp(db)
	if err != nil {
		panic(err)
	}

	// Start the application.
	application.Run()
}
