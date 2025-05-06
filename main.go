package main

import (
	"log"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	log.Fatal(server.ListenAndServe())
}
