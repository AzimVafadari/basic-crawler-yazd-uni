package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
		Timeout: 60 * time.Second, // Avoid hanging requests
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

	// --- START: MODIFIED LOGIC ---
	// Save to the database
	if err := f.repo.SavePage(url, htmlContent, resp.StatusCode); err != nil {
		// Check if this is the "UNIQUE" error.
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			// This is not a real error. It just means we've already
			// saved this page. We can safely ignore it and proceed.
			log.Printf("Already in DB: %s", url)
		} else {
			// This is a *real* error (e.g., database file locked)
			// In this case, we MUST return the error.
			return "", fmt.Errorf("failed to save page %s: %w", url, err)
		}
	} else {
		// This is a new page, log the success.
		log.Printf("Saved new page: %s (status %d)", url, resp.StatusCode)
	}
	// --- END: MODIFIED LOGIC ---

	// In all cases (new page or duplicate page), we return the
	// HTML content so the crawler can parse it for new links.
	return htmlContent, nil
}
