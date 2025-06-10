package tools

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestWebSearchRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"items":[{"title":"Result 1","link":"https://example.com","snippet":"Snippet"}]}`)
	}))
	defer srv.Close()

	os.Setenv("GOOGLE_SEARCH_API_KEY", "test")
	os.Setenv("GOOGLE_SEARCH_CX", "cx")
	os.Setenv("GOOGLE_SEARCH_ENDPOINT", srv.URL)
	defer os.Unsetenv("GOOGLE_SEARCH_API_KEY")
	defer os.Unsetenv("GOOGLE_SEARCH_CX")
	defer os.Unsetenv("GOOGLE_SEARCH_ENDPOINT")

	ws := webSearch{Query: "golang", NumResults: 1}
	res := ws.Run()

	response, ok := res.(SearchResponse)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if response.Error != "" {
		t.Fatalf("unexpected error: %s", response.Error)
	}
	if len(response.Results) != 1 || response.Results[0].Title != "Result 1" {
		t.Fatalf("unexpected results: %+v", response.Results)
	}
}

func TestWebSearchRun_EmptyQuery(t *testing.T) {
	ws := webSearch{Query: "", NumResults: 1}
	res := ws.Run()

	response, ok := res.(SearchResponse)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if response.Error != "query cannot be empty" {
		t.Fatalf("expected empty query error, got: %s", response.Error)
	}
}

func TestWebSearchRun_WhitespaceQuery(t *testing.T) {
	ws := webSearch{Query: "   \t\n  ", NumResults: 1}
	res := ws.Run()

	response, ok := res.(SearchResponse)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if response.Error != "query cannot be empty" {
		t.Fatalf("expected empty query error, got: %s", response.Error)
	}
}

func TestWebSearchRun_MissingAPIKey(t *testing.T) {
	os.Unsetenv("GOOGLE_SEARCH_API_KEY")
	os.Setenv("GOOGLE_SEARCH_CX", "cx")
	defer os.Unsetenv("GOOGLE_SEARCH_CX")

	ws := webSearch{Query: "golang", NumResults: 1}
	res := ws.Run()

	response, ok := res.(SearchResponse)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if response.Error != "GOOGLE_SEARCH_API_KEY environment variable is required" {
		t.Fatalf("expected API key error, got: %s", response.Error)
	}
}

func TestWebSearchRun_MissingSearchCX(t *testing.T) {
	os.Setenv("GOOGLE_SEARCH_API_KEY", "test")
	os.Unsetenv("GOOGLE_SEARCH_CX")
	defer os.Unsetenv("GOOGLE_SEARCH_API_KEY")

	ws := webSearch{Query: "golang", NumResults: 1}
	res := ws.Run()

	response, ok := res.(SearchResponse)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if response.Error != "GOOGLE_SEARCH_CX environment variable is required" {
		t.Fatalf("expected CX error, got: %s", response.Error)
	}
}

func TestWebSearchRun_DefaultNumResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("num") != "3" {
			t.Errorf("expected num=3, got num=%s", r.URL.Query().Get("num"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"items":[]}`)
	}))
	defer srv.Close()

	os.Setenv("GOOGLE_SEARCH_API_KEY", "test")
	os.Setenv("GOOGLE_SEARCH_CX", "cx")
	os.Setenv("GOOGLE_SEARCH_ENDPOINT", srv.URL)
	defer os.Unsetenv("GOOGLE_SEARCH_API_KEY")
	defer os.Unsetenv("GOOGLE_SEARCH_CX")
	defer os.Unsetenv("GOOGLE_SEARCH_ENDPOINT")

	// Test with NumResults = 0
	ws := webSearch{Query: "golang", NumResults: 0}
	ws.Run()

	// Test with NumResults < 0
	ws = webSearch{Query: "golang", NumResults: -1}
	ws.Run()
}

func TestWebSearchRun_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	os.Setenv("GOOGLE_SEARCH_API_KEY", "test")
	os.Setenv("GOOGLE_SEARCH_CX", "cx")
	os.Setenv("GOOGLE_SEARCH_ENDPOINT", srv.URL)
	defer os.Unsetenv("GOOGLE_SEARCH_API_KEY")
	defer os.Unsetenv("GOOGLE_SEARCH_CX")
	defer os.Unsetenv("GOOGLE_SEARCH_ENDPOINT")

	ws := webSearch{Query: "golang", NumResults: 1}
	res := ws.Run()

	response, ok := res.(SearchResponse)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if response.Error == "" {
		t.Fatalf("expected error for HTTP 500, got none")
	}
}
