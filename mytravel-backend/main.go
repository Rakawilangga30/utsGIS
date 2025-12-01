package main

import (
	"log"
	"net/http"
	"os"
	
	// Pastikan module path ini sesuai dengan di go.mod Anda
	"MyTravel/api"
)

func main() {
	// 1. Panggil koneksi DB (ini memuat .env di lokal)
	api.ConnectDB()

	// 2. Setup Routing
	http.Handle("/", api.Handler())

	// 3. Setup Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 4. Start Server
	log.Println("MyTravel backend running on port:", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}