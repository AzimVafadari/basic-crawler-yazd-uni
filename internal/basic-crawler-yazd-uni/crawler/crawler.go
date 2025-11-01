package crawler

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/fetcher"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/parser"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/uniqueness"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/urlhelper"
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

	// State
	queue []string
	mu     sync.Mutex // protects queue
}

// NewCrawler creates and initialize a new crawler.
func NewCrawler(startURLStr string, f *fetcher.Fetcher, c *uniqueness.Checker, limit int) (*Crawler, error) {
	u, err := url.Parse(startURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	domain := u.Hostname()

	return &Crawler{
		fetcher:      f,
		checker:      c,
		startURL:     u,
		targetDomain: domain,
		pageLimit:    limit,
		queue:        []string{startURLStr},
	}, nil
}

// Run starts the main crawling loop with concurrency and queue.
func (c *Crawler) Run() {
	log.Printf("Starting crawl. Target domain: %s, Page limit: %d", c.targetDomain, c.pageLimit)

	urlChan := make(chan string, 200)
	var wg sync.WaitGroup
	var pagesCrawled int32

	// --- Worker goroutine function ---
	worker := func(id int) {
		defer wg.Done()
		for currentURL := range urlChan {
			// Stop condition
			if atomic.LoadInt32(&pagesCrawled) >= int32(c.pageLimit) {
				return
			}

			htmlContent, err := c.fetcher.Fetch(currentURL)
			if err != nil {
				log.Printf("[Worker %d] ERROR: Could not fetch %s: %v", id, currentURL, err)
				continue
			}

			count := atomic.AddInt32(&pagesCrawled, 1)
			log.Printf("[Worker %d] Crawled (%d/%d): %s", id, count, c.pageLimit, currentURL)

			// Parse links
			links, err := parser.ParseLinks(currentURL, htmlContent)
			if err != nil {
				log.Printf("[Worker %d] ERROR parsing %s: %v", id, currentURL, err)
				continue
			}

			// Add discovered links to the queue
			for _, linkStr := range links {
				linkURL, err := url.Parse(linkStr)
				if err != nil {
					continue
				}
				linkURL.Fragment = ""
				cleanLinkStr := linkURL.String()

				if urlhelper.IsOnDomain(c.targetDomain, linkURL) && c.checker.IsUnique(cleanLinkStr) {
					c.mu.Lock()
					c.queue = append(c.queue, cleanLinkStr)
					c.mu.Unlock()
				}
			}
		}
	}

	// --- Dispatcher goroutine ---
	// Feeds URLs from the queue into the channel as workers process them.
	dispatcher := func() {
		for {
			// Stop when limit reached
			if atomic.LoadInt32(&pagesCrawled) >= int32(c.pageLimit) {
				close(urlChan)
				return
			}

			c.mu.Lock()
			if len(c.queue) == 0 {
				c.mu.Unlock()
				continue
			}
			// Dequeue next URL
			currentURL := c.queue[0]
			c.queue = c.queue[1:]
			c.mu.Unlock()

			select {
			case urlChan <- currentURL:
				// Sent successfully
			default:
				// Channel full → back off slightly
			}
		}
	}

	// --- Start the dispatcher ---
	go dispatcher()

	// --- Start workers ---
	workerCount := 3
	wg.Add(workerCount)
	for i := 1; i <= workerCount; i++ {
		go worker(i)
	}

	// --- Seed the queue with the start URL ---
	// (already done in constructor, but ensures first dispatch happens)
	c.mu.Lock()
	if len(c.queue) == 0 {
		c.queue = append(c.queue, c.startURL.String())
	}
	c.mu.Unlock()

	wg.Wait()
	log.Printf("Crawling finished. Crawled %d pages. Total unique URLs found: %d", pagesCrawled, c.checker.Count())
}
