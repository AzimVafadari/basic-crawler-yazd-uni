package crawler

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/fetcher"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/parser"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/uniqueness"
	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/urlhelper"
)

// Crawler manages the crawling process and statistics.
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
	mu    sync.Mutex // protects queue

	// --- Statistics ---
	startTime    time.Time
	bytesFetched int64
	errorCount   int64
	discovered   int64
}

// NewCrawler creates and initializes a new crawler instance.
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

// Run starts the main crawling loop with concurrency and collects statistics.
func (c *Crawler) Run() {
	c.startTime = time.Now()
	log.Printf("Starting crawl. Target domain: %s, Page limit: %d", c.targetDomain, c.pageLimit)

	urlChan := make(chan string, 200)
	var wg sync.WaitGroup
	var pagesCrawled int32

	// --- Rate Limiter (global delay: one request every 2 seconds) ---
	rateLimiter := time.NewTicker(2 * time.Second)
	defer rateLimiter.Stop()

	// --- Worker goroutine ---
	worker := func(id int) {
		defer wg.Done()
		for currentURL := range urlChan {
			// Stop condition
			if atomic.LoadInt32(&pagesCrawled) >= int32(c.pageLimit) {
				return
			}

			// --- Wait for permission to send next request ---
			// Global delay (shared among all workers)
			<-rateLimiter.C

			// (Alternative: delay per worker)
			// time.Sleep(2 * time.Second)

			htmlContent, err := c.fetcher.Fetch(currentURL)
			if err != nil {
				log.Printf("[Worker %d] ERROR: Could not fetch %s: %v", id, currentURL, err)
				atomic.AddInt64(&c.errorCount, 1)
				continue
			}

			// Count fetched bytes
			atomic.AddInt64(&c.bytesFetched, int64(len(htmlContent)))

			count := atomic.AddInt32(&pagesCrawled, 1)
			log.Printf("[Worker %d] 🕸 Crawled (%d/%d): %s", id, count, c.pageLimit, currentURL)

			// Parse links
			links, err := parser.ParseLinks(currentURL, htmlContent)
			if err != nil {
				log.Printf("[Worker %d] ERROR parsing %s: %v", id, currentURL, err)
				atomic.AddInt64(&c.errorCount, 1)
				continue
			}

			// Count discovered links (including duplicates)
			atomic.AddInt64(&c.discovered, int64(len(links)))

			// Add discovered links to queue
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
	dispatcher := func() {
		for {
			if atomic.LoadInt32(&pagesCrawled) >= int32(c.pageLimit) {
				close(urlChan)
				return
			}

			c.mu.Lock()
			if len(c.queue) == 0 {
				c.mu.Unlock()
				time.Sleep(500 * time.Millisecond) // prevent busy waiting
				continue
			}
			currentURL := c.queue[0]
			c.queue = c.queue[1:]
			c.mu.Unlock()

			select {
			case urlChan <- currentURL:
				// sent successfully
			default:
				// channel full → small backoff
				time.Sleep(200 * time.Millisecond)
			}
		}
	}

	// Start dispatcher
	go dispatcher()

	// Start workers
	workerCount := 3
	wg.Add(workerCount)
	for i := 1; i <= workerCount; i++ {
		go worker(i)
	}

	// Ensure queue has seed URL
	c.mu.Lock()
	if len(c.queue) == 0 {
		c.queue = append(c.queue, c.startURL.String())
	}
	c.mu.Unlock()

	wg.Wait()

	// --- Print statistics ---
	duration := time.Since(c.startTime)
	c.mu.Lock()
	queueLength := len(c.queue)
	c.mu.Unlock()

	log.Println("=== Crawl Statistics ===")
	log.Printf("⏱  Duration: %v", duration)
	log.Printf("📦 Total bytes fetched: %d bytes", c.bytesFetched)
	log.Printf("❌ Errors encountered: %d", c.errorCount)
	log.Printf("🔗 Links discovered (including duplicates): %d", c.discovered)
	log.Printf("🧾 Queue length at finish: %d", queueLength)
	log.Printf("✅ Unique URLs stored: %d", c.checker.Count())
	log.Printf("Crawling finished. Crawled %d pages.", pagesCrawled)
}
