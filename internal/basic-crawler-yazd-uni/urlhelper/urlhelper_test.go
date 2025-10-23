package urlhelper

import (
	"net/url"
	"testing" // Import the special testing package
)

// TestResolveURL is our test function. Go will run this automatically.
func TestResolveURL(t *testing.T) {
	// 1. ARRANGE (Setup)
	// We need a base URL for all our tests.
	// We use t.Fatal() here because if this setup fails,
	// the test itself is broken and we must stop.
	baseURL, err := url.Parse("https://example.com/products/main.html")
	if err != nil {
		t.Fatalf("Failed to parse base URL: %v", err)
	}

	// This is a "table-driven test". We define a list of test cases.
	// This is a very common and powerful pattern in Go.
	testCases := []struct {
		name     string // A name for this specific test case
		href     string // The input href to test
		expected string // The expected absolute URL string
	}{
		{
			name:     "Relative path from root",
			href:     "/about-us",
			expected: "https://example.com/about-us",
		},
		{
			name:     "Relative path from current dir",
			href:     "item/123",
			expected: "https://example.com/products/item/123",
		},
		{
			name:     "Relative path going up",
			href:     "../support",
			expected: "https://example.com/support",
		},
		{
			name:     "Fully qualified URL (should not change)",
			href:     "https://www.google.com",
			expected: "https://www.google.com",
		},
	}

	// 2. ACT & 3. ASSERT (Loop through all our test cases)
	for _, tc := range testCases {
		// t.Run() creates a sub-test, which gives a nice, clean
		// output in the test runner.
		t.Run(tc.name, func(t *testing.T) {
			// ACT: Call the function we want to test
			actualURL := ResolveURL(baseURL, tc.href)

			// ASSERT: Check the result
			if actualURL == nil {
				t.Errorf("ResolveURL returned nil, but expected %s", tc.expected)
				return // Stop this sub-test
			}

			actualStr := actualURL.String()
			if actualStr != tc.expected {
				// This is how you fail a test.
				// t.Errorf logs an error message but continues running other tests.
				t.Errorf("Failed: expected %s, but got %s", tc.expected, actualStr)
			}
		})
	}
}
