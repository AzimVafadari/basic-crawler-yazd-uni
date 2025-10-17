package main

import (
	"go-crawler/storage"
	"log"
)

func main() {
	// Initialize the database repository.
	// This will create a file named "crawler.db" in the project root.
	repo, err := storage.NewRepository("crawler.db")
	if err != nil {
		log.Fatalf("FATAL: Could not initialize storage: %v", err)
	}
	// Defer closing the database until the main function exits.
	defer repo.Close()

	// ... The rest of your crawler initialization will go here ...
	// You will pass the `repo` object to your crawler so it can save pages.
	log.Println("Starting crawler...")
}
