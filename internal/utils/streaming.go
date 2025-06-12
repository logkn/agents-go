package utils

import "strings"

// Returns true iff tokens is a valid XML opening or closing tag.
func IsXML(tokens string) bool {
	tokens = strings.TrimSpace(tokens)
	if len(tokens) < 3 {
		return false
	}

	// Check for opening tag: <tagname> or <tagname attr="value">
	if tokens[0] == '<' && tokens[len(tokens)-1] == '>' {
		inner := tokens[1 : len(tokens)-1]

		// Check for closing tag: </tagname>
		if len(inner) > 0 && inner[0] == '/' {
			tagName := strings.TrimSpace(inner[1:])
			// For closing tags, the tag name should not have leading/trailing spaces in the original
			if tagName != inner[1:] {
				return false // Had spaces, invalid
			}
			return isValidTagName(tagName)
		}

		// Check for self-closing tag: <tagname/>
		if len(inner) > 0 && inner[len(inner)-1] == '/' {
			tagContent := strings.TrimSpace(inner[:len(inner)-1])
			// Check if trimming changed the content (which means it had spaces)
			if tagContent != inner[:len(inner)-1] {
				// Has spaces, only valid if it's a proper attribute structure
				parts := strings.Fields(inner[:len(inner)-1])
				if len(parts) > 0 {
					return isValidTagName(parts[0])
				}
				return false
			}
			return isValidTagName(tagContent)
		}

		// Check for opening tag: <tagname> or <tagname attr="value">
		trimmed := strings.TrimSpace(inner)
		if trimmed != inner {
			// Has leading/trailing spaces, only valid if it's attributes (multiple parts)
			parts := strings.Fields(inner)
			if len(parts) > 1 {
				// Multiple parts means it has attributes, validate tag name
				return isValidTagName(parts[0])
			}
			// Single part with spaces around it is invalid XML
			return false
		}

		// No leading/trailing spaces, check if it's just a tag name
		if !strings.ContainsAny(inner, " \t\n\r") {
			return isValidTagName(inner)
		}

		// Has internal spaces, must be attributes
		parts := strings.Fields(inner)
		if len(parts) > 0 {
			return isValidTagName(parts[0])
		}
	}

	return false
}

// isValidTagName checks if a string is a valid XML tag name
func isValidTagName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// XML tag names must start with a letter or underscore
	first := name[0]
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' {
		return false
	}

	// Rest can be letters, digits, hyphens, periods, or underscores
	for i := 1; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '.' && c != '_' {
			return false
		}
	}

	return true
}

// Yields the same stream, but with XML opening and closing tags
// grouped as one string part.
func GroupXML(stream chan string) chan string {
	output := make(chan string)

	go func() {
		defer close(output)

		var buffer strings.Builder
		var tagBuffer strings.Builder
		var inTag bool

		for chunk := range stream {
			for _, char := range chunk {
				if inTag {
					// We're inside a tag, accumulate until we see '>'
					tagBuffer.WriteRune(char)
					if char == '>' {
						// Tag complete, check if we have non-tag content to send first
						if buffer.Len() > 0 {
							output <- buffer.String()
							buffer.Reset()
						}
						// Send the complete tag
						output <- tagBuffer.String()
						tagBuffer.Reset()
						inTag = false
					}
				} else {
					if char == '<' {
						// Starting a tag, send any accumulated content first
						if buffer.Len() > 0 {
							output <- buffer.String()
							buffer.Reset()
						}
						// Start accumulating the tag
						tagBuffer.WriteRune(char)
						inTag = true
					} else {
						// Regular content, accumulate
						buffer.WriteRune(char)
					}
				}
			}
		}

		// Send any remaining content
		if buffer.Len() > 0 {
			output <- buffer.String()
		}
		if tagBuffer.Len() > 0 {
			output <- tagBuffer.String()
		}
	}()

	return output
}

type Token struct {
	Content    string
	IsThinking bool
}

func DetectThinking(stream chan string) chan Token {
	output := make(chan Token)

	go func() {
		defer close(output)

		isThinking := false
		for chunk := range stream {
			switch chunk {
			case "<think>":
				isThinking = true

			case "</think>":
				isThinking = false
			case "":
				continue // Skip empty chunks
			default:
				output <- Token{
					Content:    chunk,
					IsThinking: isThinking,
				}
			}
		}
	}()
	return output
}
