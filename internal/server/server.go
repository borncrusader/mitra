package server

import (
	"fmt"
	"log"
	"net/http"

	"mitra/internal/proto"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func Start(cfg *proto.Config) error {
	http.HandleFunc("/", helloHandler)

	log.Printf("Server starting on http://localhost%s", cfg.Server.Port)

	return http.ListenAndServe(cfg.Server.Port, nil)
}
