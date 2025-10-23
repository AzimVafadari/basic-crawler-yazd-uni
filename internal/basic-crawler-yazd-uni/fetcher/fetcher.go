package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/storage"
)

// Fetcher is responsible for downloading web pages
// and saving them into the storage repository.
type Fetcher struct {
	repo *storage.Repository
}

// NewFetcher creates and returns a new Fetcher instance.
func NewFetcher(repo *storage.Repository) *Fetcher {
	return &Fetcher{repo: repo}
}

// Fetch downloads a web page from the given URL and returns its HTML content.
// It also stores the page details (URL, HTML, status code, timestamp) in the database.
// Returns the HTML content as a string if successful, or an error otherwise.
func (f *Fetcher) Fetch(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body from %s: %w", url, err)
	}

	if err := f.repo.SavePage(url, string(body), resp.StatusCode); err != nil {
		return "", fmt.Errorf("failed to save page %s: %w", url, err)
	}

	return string(body), nil
}
