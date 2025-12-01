package main

import (
	"log"
	"net/http"
	"os"

	// Import folder "api", tapi nama packagenya sekarang "handler"
	"MyTravel/api" 
)

func main() {
	// SetupRouter dari package handler (karena di api/api.go package handler)
	r := handler.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("MyTravel backend running on port:", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}