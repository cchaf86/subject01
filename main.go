package main

import (
	"log"
	"net/http"

	"test/golang/app"
)

func main() {
	db := app.MustSetupDB("data.db")
	server := app.NewServer(db)

	mux := http.NewServeMux()
	app.RegisterRoutes(mux, server)

	log.Println("server listening on :8081")
	if err := http.ListenAndServe(":8081", app.WithCORS(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
