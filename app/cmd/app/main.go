package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Application successfully started")
	mux := http.NewServeMux()

	port := ":" + "8080"
	log.Printf("Server is running on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}

}
