package main

import (
	"log"
	"net/http"

	"github.com/serjyuriev/diploma-1/internal/app"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			http.ListenAndServe(config.GetConfig().RunAddress, http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
				},
			))
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
