package utils

import (
	"slices"
	"testing"
)

func TestIsXML(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"<tag>", true},
		{"</tag>", true},
		{"<tag/>", true},
		{"<tag attr='value'>", true},
		{"<valid_name>", true},
		{"<valid-name>", true},
		{"<valid.name>", true},
		{"<_underscore>", true},
		{"not xml", false},
		{"<>", false},
		{"<123invalid>", false},
		{"< space >", false},
		{"<incomplete", false},
		{"incomplete>", false},
		{"", false},
		{"<a>", true},
		{"</a>", true},
	}

	for _, test := range tests {
		result := IsXML(test.input)
		if result != test.expected {
			t.Errorf("IsXML(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestGroupXML(t *testing.T) {
	// Create input channel
	input := make(chan string, 10)

	// Send test tokens
	tokens := []string{"Hello ", "<tag", ">", "content", "</tag>", " world"}
	go func() {
		defer close(input)
		for _, token := range tokens {
			input <- token
		}
	}()

	// Process with GroupXML
	output := GroupXML(input)

	// Collect results
	var results []string
	for result := range output {
		results = append(results, result)
	}

	// Verify results contain grouped XML
	found := slices.Contains(results, "<tag>")

	if !found {
		t.Errorf("Expected to find '<tag>' as a grouped result, got: %v", results)
	}
}

func TestGroupXMLCharByChar(t *testing.T) {
	// Create input channel
	input := make(chan string, 20)

	// Send character-by-character tokens to test the new implementation
	tokens := []string{"H", "e", "l", "l", "o", " ", "<", "t", "a", "g", ">", "c", "o", "n", "t", "e", "n", "t", "<", "/", "t", "a", "g", ">", " ", "w", "o", "r", "l", "d"}
	go func() {
		defer close(input)
		for _, token := range tokens {
			input <- token
		}
	}()

	// Process with GroupXML
	output := GroupXML(input)

	// Collect results
	var results []string
	for result := range output {
		results = append(results, result)
	}

	// Verify we get complete tags and content
	expectedTags := []string{"<tag>", "</tag>"}
	for _, expectedTag := range expectedTags {
		found := slices.Contains(results, expectedTag)
		if !found {
			t.Errorf("Expected to find '%s' as a grouped result, got: %v", expectedTag, results)
		}
	}

	// Should also have "Hello " and "content" and " world" as separate results
	expectedContent := []string{"Hello ", "content", " world"}
	for _, expectedText := range expectedContent {
		found := slices.Contains(results, expectedText)
		if !found {
			t.Errorf("Expected to find '%s' as content, got: %v", expectedText, results)
		}
	}
}
