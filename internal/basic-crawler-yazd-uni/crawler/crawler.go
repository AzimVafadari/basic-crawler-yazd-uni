package crawler

import (
	"fmt"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/fetcher"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/parser"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/uniqueness"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/urlhelper"

	"log"
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

// Run starts the main crawling loop.
func (c *Crawler) Run() {
	log.Printf("Starting crawl. Target domain: %s, Page limit: %d", c.targetDomain, c.pageLimit)

	pagesCrawled := 0

	for len(c.queue) > 0 && pagesCrawled < c.pageLimit {

		// --- 1. DEQUEUE ---
		currentURL := c.queue[0]
		c.queue = c.queue[1:]

		// --- 2. REMOVE THE UNIQUENESS CHECK ---
		// The fetcher now handles duplicates, so we
		// no longer need to check here.
		/*
			if !c.checker.IsUnique(currentURL) {
				log.Printf("Skipping already seen URL: %s", currentURL)
				continue
			}
		*/

		// --- 3. FETCH ---
		// The fetcher will now return HTML even for duplicates,
		// and won't return a "UNIQUE" error.
		htmlContent, err := c.fetcher.Fetch(currentURL)
		if err != nil {
			// This will only be for *real* errors now.
			log.Printf("ERROR: Could not fetch %s: %v", currentURL, err)
			continue
		}
		pagesCrawled++ // We successfully fetched and processed a page

		// --- 4. PARSE ---
		links, err := parser.ParseLinks(currentURL, htmlContent)
		if err != nil {
			log.Printf("ERROR: Could not parse links on %s: %v", currentURL, err)
			continue
		}
		log.Printf("Found %d new links on %s", len(links), currentURL)

		// --- 5. PROCESS & ENQUEUE NEW LINKS ---
		for _, linkStr := range links {
			linkURL, err := url.Parse(linkStr)
			if err != nil {
				log.Printf("WARN: Found invalid link '%s': %v", linkStr, err)
				continue
			}

			linkURL.Fragment = "" // Clean the link
			cleanLinkStr := linkURL.String()

			// --- THIS IS THE NEW HOME FOR THE UNIQUENESS CHECK ---
			// We only add new links to the queue *if* they are:
			// 1. On our target domain
			// 2. Not already seen by our checker
			if urlhelper.IsOnDomain(c.targetDomain, linkURL) && c.checker.IsUnique(cleanLinkStr) {
				// It's a new, valid link. Add it to the queue!
				c.queue = append(c.queue, cleanLinkStr)
			}
		}
	}

	log.Printf("Crawling finished. Crawled %d pages. Total unique URLs found: %d", pagesCrawled, c.checker.Count())
}
