package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	storage, err := NewPostgresStorage()

	if err != nil {
		log.Fatal(err)
	}

	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer("127.0.0.1:3000", storage)

	server.Run()
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
