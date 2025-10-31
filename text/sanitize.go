package text

import (
	"strings"

	"github.com/forPelevin/gomoji"
)

// CleanStringFromEmojis removes all emoji characters from the input string and reduces consecutive spaces to a single space.
func CleanStringFromEmojis(s string) string {
	s = gomoji.RemoveEmojis(s)
	s = strings.Join(strings.Fields(s), " ")

	return s
}
