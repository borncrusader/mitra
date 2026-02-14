package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", helloHandler)

	log.Printf("Server starting on http://localhost%s", config.Server.Port)

	if err := http.ListenAndServe(config.Server.Port, nil); err != nil {
		log.Fatal(err)
	}
}
