package ai

import (
	"testing"
)

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Hey Rufus! How's it going?",
			expected: "hey rufus! how's it going?",
		},
		{
			input:    "I'm not a bot, I'm just a girl talking to you.",
			expected: "i'm not a bot, i'm just a girl talking to you",
		},
		{
			input:    "no_output",
			expected: "",
		},
		{
			input:    "   arre yaar   kya scene hai?   ",
			expected: "arre yaar kya scene hai?",
		},
	}

	for _, tt := range tests {
		actual := SanitizeText(tt.input)
		if actual != tt.expected {
			t.Errorf("SanitizeText(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}
