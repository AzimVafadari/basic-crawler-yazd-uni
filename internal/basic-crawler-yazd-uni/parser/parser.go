package parser

import (
	"net/url"
	"strings"

	"github.com/AzimVafadari/basic-crawler-yazd-uni/internal/basic-crawler-yazd-uni/urlhelper"
	"golang.org/x/net/html"
)

// ParseLinks extracts and normalizes all links (<a href="...">) from HTML content.
// It returns a list of absolute URLs as strings.
func ParseLinks(base string, htmlContent string) ([]string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var links []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					resolved := urlhelper.ResolveURL(baseURL, attr.Val)
					if resolved != nil {
						links = append(links, resolved.String())
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return links, nil
}
