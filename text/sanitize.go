package text

import (
	"strings"

	"github.com/forPelevin/gomoji"
)

// NormalizeText reduces consecutive whitespace (including newline/tab) to a single space
// and trims leading/trailing spaces. This uses strings.Fields internally.
func NormalizeText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// RemoveEmojis removes all emoji characters from the input string.
func RemoveEmojis(s string) string {
	return gomoji.RemoveEmojis(s)
}

// CleanStringFromEmojis removes all emoji characters and normalizes spaces in the input string.
// This function first removes emojis, then normalizes spaces using NormalizeText.
func CleanStringFromEmojis(s string) string {
	s = RemoveEmojis(s)
	s = NormalizeText(s)
	return s
}
