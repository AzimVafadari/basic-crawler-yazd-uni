package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/storage"
)

// Fetcher handles downloading web pages and storing them in the database.
type Fetcher struct {
	repo *storage.Repository
}

// NewFetcher creates and returns a new Fetcher instance.
func NewFetcher(repo *storage.Repository) *Fetcher {
	return &Fetcher{repo: repo}
}

// Fetch downloads a web page and stores its content, status code, and timestamp
// in the SQLite database using the provided storage.Repository.
// Returns the HTML content if successful, or an error otherwise.
func (f *Fetcher) Fetch(url string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second, // Avoid hanging requests
	}

	log.Printf("Fetching: %s ...", url)

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body from %s: %w", url, err)
	}

	htmlContent := string(body)

	// Save to the database
	if err := f.repo.SavePage(url, htmlContent, resp.StatusCode); err != nil {
		return "", fmt.Errorf("failed to save page %s: %w", url, err)
	}

	log.Printf("Saved page: %s (status %d)", url, resp.StatusCode)

	return htmlContent, nil
}
