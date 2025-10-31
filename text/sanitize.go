package text

import (
	"strings"

	"github.com/forPelevin/gomoji"
)

// RemoveEmojisFromString removes all emoji characters from the input string.
func RemoveEmojisFromString(s string) string {
	return gomoji.RemoveEmojis(s)
}

// NormalizeSpaces reduces consecutive spaces to a single space and trims leading/trailing spaces.
func NormalizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// CleanStringFromEmojis removes all emoji characters and normalizes spaces in the input string.
func CleanStringFromEmojis(s string) string {
	s = RemoveEmojisFromString(s)
	s = NormalizeSpaces(s)
	return s
}
