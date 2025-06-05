package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// WebSearch performs web searches using the Google Custom Search API.
// It requires GOOGLE_SEARCH_API_KEY and GOOGLE_SEARCH_CX environment variables.
// Optionally, GOOGLE_SEARCH_ENDPOINT can be set to override the default API endpoint.
type WebSearch struct {
	// Query is the search query string (must be non-empty after trimming).
	Query string `json:"query" description:"The search query to execute"`
	// NumResults is the maximum number of results to return (defaults to 3 if <= 0).
	NumResults int `json:"num_results" description:"Maximum number of search results to return"`
}

// SearchResult represents a single search result item.
type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// SearchResponse represents the response from a web search operation.
type SearchResponse struct {
	Results []SearchResult `json:"results,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// Run performs the web search using the Google Custom Search API.
// It validates the query, checks for required environment variables,
// and returns a structured SearchResponse with results or error information.
//
// Required environment variables:
//   - GOOGLE_SEARCH_API_KEY: Google API key with Custom Search API enabled
//   - GOOGLE_SEARCH_CX: Custom Search Engine ID
//
// Optional environment variables:
//   - GOOGLE_SEARCH_ENDPOINT: Custom API endpoint (defaults to Google's API)
func (w WebSearch) Run() any {
	// Validate query
	query := strings.TrimSpace(w.Query)
	if query == "" {
		return SearchResponse{Error: "query cannot be empty"}
	}

	// Check environment variables
	apiKey := os.Getenv("GOOGLE_SEARCH_API_KEY")
	cx := os.Getenv("GOOGLE_SEARCH_CX")
	if apiKey == "" {
		return SearchResponse{Error: "GOOGLE_SEARCH_API_KEY environment variable is required"}
	}
	if cx == "" {
		return SearchResponse{Error: "GOOGLE_SEARCH_CX environment variable is required"}
	}

	// Set default number of results
	if w.NumResults <= 0 {
		w.NumResults = 3
	}

	// Build request parameters
	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("cx", cx)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", w.NumResults))

	// Determine endpoint
	endpoint := os.Getenv("GOOGLE_SEARCH_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://www.googleapis.com/customsearch/v1"
	}
	reqURL := endpoint + "?" + params.Encode()

	// Make HTTP request
	resp, err := http.Get(reqURL)
	if err != nil {
		return SearchResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return SearchResponse{Error: fmt.Sprintf("search API returned status %s", resp.Status)}
	}

	// Parse response
	var payload struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SearchResponse{Error: fmt.Sprintf("failed to parse response: %v", err)}
	}

	// Convert to SearchResult format
	results := make([]SearchResult, 0, len(payload.Items))
	for _, item := range payload.Items {
		results = append(results, SearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		})
	}

	return SearchResponse{Results: results}
}
