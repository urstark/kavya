package ai

import (
	"regexp"
	"strings"
)

var (
	spaceRegex = regexp.MustCompile(`\s+`)
)

// SanitizeText cleans up output to match modern casual texting:
// - Lowercases text for that relaxed Gen Z Telegram/WhatsApp vibe
// - Trims formal trailing periods that make texts look like corporate emails
// - Normalizes whitespace without brutally slicing off words
func SanitizeText(text string) string {
	if text == "" || text == "no_output" {
		return ""
	}

	text = strings.ToLower(text)
	text = spaceRegex.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Remove trailing period if present (Gen Z doesn't end texts with a formal period)
	text = strings.TrimSuffix(text, ".")

	return text
}
