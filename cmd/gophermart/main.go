package main

import (
	"log"
	"net/http"
	"time"

	"github.com/serjyuriev/diploma-1/internal/app"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
)

func main() {
	log.Println("hey")
	cfg := config.GetConfig()
	go func() {
		http.ListenAndServe(cfg.RunAddress, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		}))
	}()
	time.Sleep(10 * time.Minute)

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
		return
	}

	log.Println("starting application...")
	if err := app.Start(); err != nil {
		log.Printf("an error occured while the app was working: %v", err)
	}
}
