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

	// The main loop:
	// 1. Check if the queue is empty.
	// 2. Check if we have hit our page limit.
	for len(c.queue) > 0 && pagesCrawled < c.pageLimit {

		// --- 1. DEQUEUE ---
		// Get the next URL from the front of the queue
		currentURL := c.queue[0]
		c.queue = c.queue[1:] // This moves the queue forward

		// --- 2. CHECK UNIQUENESS ---
		// Use your uniqueness.Checker to see if this URL is new
		if !c.checker.IsUnique(currentURL) {
			log.Printf("Skipping already seen URL: %s", currentURL)
			continue // Go to the next iteration of the loop
		}

		// --- 3. FETCH ---
		// Use your fetcher.Fetch method
		htmlContent, err := c.fetcher.Fetch(currentURL)
		if err != nil {
			log.Printf("ERROR: Could not fetch %s: %v", currentURL, err)
			continue // Go to the next iteration
		}
		pagesCrawled++ // We successfully fetched a page

		// --- 4. PARSE ---
		// Use your parser.ParseLinks function
		links, err := parser.ParseLinks(currentURL, htmlContent)
		if err != nil {
			log.Printf("ERROR: Could not parse links on %s: %v", currentURL, err)
			continue // Go to the next iteration
		}

		// --- 5. PROCESS & ENQUEUE NEW LINKS ---
		for _, linkStr := range links {
			// Parse the new link string into a url.URL struct
			linkURL, err := url.Parse(linkStr)
			if err != nil {
				log.Printf("WARN: Found invalid link '%s': %v", linkStr, err)
				continue
			}

			// Clean the link (remove #fragments)
			linkURL.Fragment = ""
			cleanLinkStr := linkURL.String()

			// Use your NEW urlhelper.IsOnDomain function
			if urlhelper.IsOnDomain(c.targetDomain, linkURL) {
				// If it's on our domain, add it to the queue
				// The checker will handle duplicates next time it's dequeued
				c.queue = append(c.queue, cleanLinkStr)
			}
		}
	}

	log.Printf("Crawling finished. Crawled %d pages. Total unique URLs found: %d", pagesCrawled, c.checker.Count())
}
