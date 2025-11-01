package main

import (
	"log"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/crawler"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/fetcher"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/storage"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/uniqueness"
)

const (
	START_URL   = "https://torob.com/"
	DB_FILENAME = "crawler.db"
	PAGE_LIMIT  = 11000
)

func main() {
	log.Println("--- Crawler Starting ---")
	log.Printf("Target: %s, Limit: %d pages", START_URL, PAGE_LIMIT)

	// --- 1. Initialize Dependencies ---
	repo, err := storage.NewRepository(DB_FILENAME)
	if err != nil {
		log.Fatalf("FATAL: Could not initialize storage: %v", err)
	}
	defer repo.Close()

	f := fetcher.NewFetcher(repo)
	c := uniqueness.NewChecker()

	// --- 2. Hydrate State ---
	log.Println("Hydrating uniqueness checker from database...")
	existingURLs, err := repo.GetAllURLs()
	if err != nil {
		log.Fatalf("FATAL: Could not get existing URLs: %v", err)
	}
	c.Preload(existingURLs)
	log.Printf("Loaded %d existing URLs into checker.", c.Count())

	// --- 3. Initialize The Engine ---
	crawlEngine, err := crawler.NewCrawler(START_URL, f, c, PAGE_LIMIT)
	if err != nil {
		log.Fatalf("FATAL: Could not initialize crawler: %v", err)
	}

	// --- 4. Run ---
	crawlEngine.Run()

	log.Println("--- Crawler Finished ---")
}
