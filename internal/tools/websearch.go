package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// WebSearch represents parameters for performing a web search.
type WebSearch struct {
	// Query is the search query string.
	Query string `json:"query"`
	// NumResults is the maximum number of results to return.
	NumResults int `json:"num_results"`
}

// searchResult represents a single search result item.
type searchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// Run performs the web search using the Google Custom Search API.
// The API key and search engine ID (cx) must be provided via the
// GOOGLE_SEARCH_API_KEY and GOOGLE_SEARCH_CX environment variables.
func (w WebSearch) Run() any {
	apiKey := os.Getenv("GOOGLE_SEARCH_API_KEY")
	cx := os.Getenv("GOOGLE_SEARCH_CX")
	if apiKey == "" || cx == "" {
		return fmt.Sprintf("missing GOOGLE_SEARCH_API_KEY or GOOGLE_SEARCH_CX")
	}
	if w.NumResults <= 0 {
		w.NumResults = 3
	}

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("cx", cx)
	params.Set("q", w.Query)
	params.Set("num", fmt.Sprintf("%d", w.NumResults))

	endpoint := os.Getenv("GOOGLE_SEARCH_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://www.googleapis.com/customsearch/v1"
	}
	reqURL := endpoint + "?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return fmt.Sprintf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("search API returned status %s", resp.Status)
	}

	var payload struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Sprintf("decode error: %v", err)
	}

	results := make([]searchResult, 0, len(payload.Items))
	for _, it := range payload.Items {
		results = append(results, searchResult{
			Title:   it.Title,
			Link:    it.Link,
			Snippet: it.Snippet,
		})
	}
	return results
}
