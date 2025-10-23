package urlhelper

import (
	"log"
	"net/url"
)

// ResolveURL takes the page's base URL and a link string (href)
// and returns a complete, absolute URL.
// It returns 'nil' if the href is invalid.
func ResolveURL(baseURL *url.URL, href string) *url.URL {
	resolvedURL, err := baseURL.Parse(href)

	if err != nil {
		log.Printf("WARNNING: Could not parse %s based on %s: %v", href, baseURL, err)
		return nil
	}

	return resolvedURL
}
