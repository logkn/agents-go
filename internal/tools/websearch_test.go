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

	ws := WebSearch{Query: "golang", NumResults: 1}
	res := ws.Run()

	results, ok := res.([]searchResult)
	if !ok {
		t.Fatalf("unexpected type %T", res)
	}
	if len(results) != 1 || results[0].Title != "Result 1" {
		t.Fatalf("unexpected results: %+v", results)
	}
}
