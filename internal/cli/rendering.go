package cli

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var mdRenderer, _ = glamour.NewTermRenderer(
	glamour.WithStylePath("customstyle.json"),
	glamour.WithWordWrap(0),
)

type ContentSegment struct {
	Text       string
	IsThinking bool
	InProgress bool
}

func normalizeThinkTags(text string) string {
	text = strings.TrimSpace(text)
	thinkStartRe := regexp.MustCompile(`<think>\s*`)
	thinkEndRe := regexp.MustCompile(`\s*</think>`)

	if thinkStartRe.MatchString(text) && !thinkEndRe.MatchString(text) {
		text += "</think>"
	}
	return text
}

func parseContentSegments(text string, isStreaming bool) []ContentSegment {
	text = normalizeThinkTags(text)

	thinkTagRe := regexp.MustCompile(`(?s)<think>\s*(.*?)\s*</think>`)
	matches := thinkTagRe.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		return []ContentSegment{{Text: text, IsThinking: false, InProgress: isStreaming}}
	}

	var segments []ContentSegment
	lastEnd := 0

	for _, match := range matches {
		if match[0] > lastEnd {
			content := text[lastEnd:match[0]]
			if strings.TrimSpace(content) != "" {
				segments = append(segments, ContentSegment{Text: content, IsThinking: false, InProgress: false})
			}
		}

		thinkingContent := text[match[2]:match[3]]
		if strings.TrimSpace(thinkingContent) != "" {
			segments = append(segments, ContentSegment{Text: thinkingContent, IsThinking: true, InProgress: false})
		}

		lastEnd = match[1]
	}

	if lastEnd < len(text) {
		content := text[lastEnd:]
		if strings.TrimSpace(content) != "" {
			segments = append(segments, ContentSegment{Text: content, IsThinking: false, InProgress: false})
		}
	}

	// Set InProgress to true for the last segment if we're streaming
	if isStreaming && len(segments) > 0 {
		segments[len(segments)-1].InProgress = true
	}

	return segments
}

func renderContent(content string, isThinking bool, hideThoughts bool) string {
	if isThinking {

		var renderContent string
		if hideThoughts {
			renderContent = "Thinking..."
		} else {
			renderContent = content
		}
		// For thinking sections, apply gray color to all text including inline code
		// We need to override the markdown renderer's color choices
		thinkingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(gray)).Italic(true)

		// Apply the thinking style to the raw content without markdown processing
		// to ensure consistent gray coloring throughout
		return thinkingStyle.Render(renderContent)
	}

	rendered, err := mdRenderer.Render(content)
	if err != nil {
		rendered = content
	} else {
		rendered = strings.TrimSpace(rendered)
	}

	return rendered
}

func RenderMarkdown(text string, hideThoughts bool, addBullets bool, isStreaming bool, spinnerView string) string {
	segments := parseContentSegments(text, isStreaming)
	if len(segments) == 0 {
		return ""
	}

	// In hide thoughts mode, if there are any non-thinking segments, skip all thinking segments
	if hideThoughts {
		hasNonThinking := false
		for _, segment := range segments {
			if !segment.IsThinking {
				hasNonThinking = true
				break
			}
		}
		if hasNonThinking {
			// Filter out all thinking segments
			var filteredSegments []ContentSegment
			for _, segment := range segments {
				if !segment.IsThinking {
					filteredSegments = append(filteredSegments, segment)
				}
			}
			segments = filteredSegments
		}
	}

	var result strings.Builder
	for i, segment := range segments {
		if i > 0 {
			result.WriteString("\n\n")
		}

		// Add bullet point for each content segment if requested
		if addBullets {
			var bullet string
			if segment.InProgress && !segment.IsThinking {
				// Use blinking spinner for in-progress non-thinking segments
				bullet = spinnerView
			} else {
				// Use static bullet
				bullet = "● "
				if segment.IsThinking {
					bullet = lipgloss.NewStyle().Foreground(lipgloss.Color(gray)).Render(bullet)
				}
			}
			result.WriteString(bullet)
		}

		result.WriteString(renderContent(segment.Text, segment.IsThinking, hideThoughts))
	}

	// Only trim leading/trailing whitespace from the overall result, not individual lines
	// This preserves spinner spacing while cleaning up the overall output
	return strings.Trim(result.String(), "\n")
}

func truncateWithEllipsis(s string, maxLen int) string {
	if maxLen < 3 {
		return s[:maxLen] // or handle this edge case differently
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
