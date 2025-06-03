package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func JsonDumps(data map[string]any, indent int) string {
	// Convert the map to JSON with indentation
	indentString := ""
	for range indent {
		indentString += " "
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indentString)

	err := encoder.Encode(data)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	result := buf.String()
	// Remove the trailing newline that Encode adds
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Unescape unicode sequences
	result = unescapeUnicode(result)
	return result
}

func JsonDumpsObj(data any) string {
	// Convert the object to JSON with indentation

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(data)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	result := buf.String()
	// Remove the trailing newline that Encode adds
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Unescape unicode sequences
	result = unescapeUnicode(result)
	return result
}

func unescapeUnicode(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if i < len(s)-5 && s[i] == '\\' && s[i+1] == 'u' {
			// Parse the 4-digit hex code
			hexCode := s[i+2 : i+6]
			if codePoint, err := strconv.ParseInt(hexCode, 16, 32); err == nil {
				result.WriteRune(rune(codePoint))
				i += 5 // Skip the entire \uXXXX sequence
			} else {
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}
