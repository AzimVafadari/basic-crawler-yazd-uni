package uniqueness

import (
	"sync"
)

// Checker is responsible for ensuring that each URL is processed only once.
// It keeps an in-memory record of all seen URLs using a thread-safe map.
type Checker struct {
	mu   sync.RWMutex
	seen map[string]struct{}
}

// NewChecker creates and returns a new instance of the Checker.
func NewChecker() *Checker {
	return &Checker{
		seen: make(map[string]struct{}),
	}
}

// IsUnique checks whether the given URL has been seen before.
// If the URL is new, it marks it as seen and returns true.
// If it already exists, it returns false.
func (c *Checker) IsUnique(url string) bool {
	// Check if the URL has already been seen
	c.mu.RLock()
	_, exists := c.seen[url]
	c.mu.RUnlock()

	if exists {
		return false
	}

	// Mark as seen
	c.mu.Lock()
	c.seen[url] = struct{}{}
	c.mu.Unlock()

	return true
}

// FilterUnique takes a list of URLs and returns only those that have not
// been seen before. All returned URLs will be marked as seen automatically.
func (c *Checker) FilterUnique(urls []string) []string {
	var unique []string

	for _, u := range urls {
		if c.IsUnique(u) {
			unique = append(unique, u)
		}
	}

	return unique
}

// Count returns the total number of unique URLs that have been recorded so far.
func (c *Checker) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.seen)
}

// Reset clears all stored URLs from the Checker.
// This can be useful between crawl sessions or for testing purposes.
func (c *Checker) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen = make(map[string]struct{})
}

// Preload populates the checker with a list of already-seen URLs.
// This is used to hydrate the checker from the database on startup.
func (c *Checker) Preload(urls []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, url := range urls {
		c.seen[url] = struct{}{} // Add each URL to the 'seen' map
	}
}
