package urlhelper

import (
	"log"
	"net/url"
	"strings"
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

// IsOnDomain checks if a given URL belongs to a target domain or one of its subdomains.
func IsOnDomain(targetDomain string, checkURL *url.URL) bool {
	hostName := checkURL.Hostname()

	// Check for an exact match or if it's a subdomain.
	return strings.HasSuffix(hostName, "."+targetDomain) || hostName == targetDomain
}
