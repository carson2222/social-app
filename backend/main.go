package main

import (
	"log"

	"github.com/carson2222/social-app/api"
	"github.com/carson2222/social-app/storage"
)

func main() {
	// DEV
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	storage, err := storage.NewPostgresStorage()

	if err != nil {
		log.Fatal(err)
	}

	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	server := api.NewAPIServer("127.0.0.1:3000", storage)

	server.Run()
}
