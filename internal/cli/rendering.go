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

func parseContentSegments(text string) []ContentSegment {
	text = normalizeThinkTags(text)

	thinkTagRe := regexp.MustCompile(`(?s)<think>\s*(.*?)\s*</think>`)
	matches := thinkTagRe.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		return []ContentSegment{{Text: text, IsThinking: false}}
	}

	var segments []ContentSegment
	lastEnd := 0

	for _, match := range matches {
		if match[0] > lastEnd {
			content := text[lastEnd:match[0]]
			if strings.TrimSpace(content) != "" {
				segments = append(segments, ContentSegment{Text: content, IsThinking: false})
			}
		}

		thinkingContent := text[match[2]:match[3]]
		if strings.TrimSpace(thinkingContent) != "" {
			segments = append(segments, ContentSegment{Text: thinkingContent, IsThinking: true})
		}

		lastEnd = match[1]
	}

	if lastEnd < len(text) {
		content := text[lastEnd:]
		if strings.TrimSpace(content) != "" {
			segments = append(segments, ContentSegment{Text: content, IsThinking: false})
		}
	}

	return segments
}

func renderContent(content string, isThinking bool) string {
	if mdRenderer == nil {
		if isThinking {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(gray)).Italic(true).Render(content)
		}
		return content
	}

	if isThinking {
		// For thinking sections, apply gray color to all text including inline code
		// We need to override the markdown renderer's color choices
		thinkingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(gray)).Italic(true)
		
		// Apply the thinking style to the raw content without markdown processing
		// to ensure consistent gray coloring throughout
		return thinkingStyle.Render(content)
	}

	rendered, err := mdRenderer.Render(content)
	if err != nil {
		rendered = content
	} else {
		rendered = strings.TrimSpace(rendered)
	}

	return rendered
}

func RenderMarkdown(text string) string {
	segments := parseContentSegments(text)
	if len(segments) == 0 {
		return ""
	}

	var result strings.Builder
	for i, segment := range segments {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(renderContent(segment.Text, segment.IsThinking))
	}

	return strings.TrimSpace(result.String())
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
