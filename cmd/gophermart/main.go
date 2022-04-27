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
	log.Println("creating application...")
	app, err := app.NewApp()
	if err != nil {
		log.Printf("unable to initialized new app: %v", err)
	}

	log.Println("starting application...")
	if err := app.Start(); err != nil {
		log.Printf("an error occured while the app was working: %v", err)
	}
}
