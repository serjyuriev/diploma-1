package main

import (
	"log"

	"github.com/serjyuriev/diploma-1/internal/app"
)

func main() {
	app, err := app.NewApp()
	if err != nil {
		log.Fatalf("unable to initialized new app: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Fatalf("an error occured while the app was working: %v", err)
	}
}
