package main

import (
	"log"

	"github.com/serjyuriev/diploma-1/internal/app/gophermart"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	app, err := gophermart.NewApp()
	if err != nil {
		log.Printf("unable to initialized new app: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Printf("an error occured while the app was working: %v", err)
	}
}
