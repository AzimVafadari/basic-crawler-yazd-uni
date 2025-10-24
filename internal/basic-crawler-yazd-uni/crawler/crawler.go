package crawler

import (
	"fmt"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/fetcher"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/uniqueness"
	"net/url"
)

// Crawler struct that manages the process
type Crawler struct {
	// Tools (Dependencies)
	fetcher *fetcher.Fetcher
	checker *uniqueness.Checker

	// Rules
	startURL     *url.URL
	targetDomain string
	pageLimit    int

	// State (WM)
	queue []string
}

// NewCrawler creates and initialize a new crawler.
func NewCrawler(startURLStr string, f *fetcher.Fetcher, c *uniqueness.Checker, limit int) (*Crawler, error) {
	// Parse the starting URL string to url.URL struct
	u, err := url.Parse(startURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	// The target domain
	domain := u.Hostname()

	// Initialize the crawler
	return &Crawler{
		fetcher:      f,
		checker:      c,
		startURL:     u,
		targetDomain: domain,
		pageLimit:    limit,
		queue:        []string{startURLStr},
	}, nil
}
