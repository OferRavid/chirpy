package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	// const logopath = "./assets/logo.png"

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/", http.FileServer(http.Dir(".")))
	// mux.Handle("/assets/logo.png", http.FileServer(http.Dir(logopath)))

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
