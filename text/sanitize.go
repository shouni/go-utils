package text

import (
	"strings"

	"github.com/forPelevin/gomoji"
)

// RemoveEmojisFromString removes all emoji characters from the input string.
// It uses the github.com/forPelevin/gomoji library for emoji detection and removal.
func RemoveEmojisFromString(s string) string {
	return gomoji.RemoveEmojis(s)
}

// NormalizeSpaces reduces consecutive spaces to a single space and trims leading/trailing spaces.
// It achieves this by splitting the string by whitespace and then joining with a single space.
func NormalizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// CleanStringFromEmojis removes all emoji characters and normalizes spaces in the input string.
// This function first removes emojis using RemoveEmojisFromString, then normalizes spaces using NormalizeSpaces.
func CleanStringFromEmojis(s string) string {
	s = RemoveEmojisFromString(s)
	s = NormalizeSpaces(s)
	return s
}
