package main

import (
	"log"
	"time"

	"github.com/serjyuriev/diploma-1/internal/app"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			time.Sleep(10 * time.Minute)
		}
	}()
	app, err := app.NewApp()
	if err != nil {
		log.Fatalf("unable to initialized new app: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Fatalf("an error occured while the app was working: %v", err)
	}
}
