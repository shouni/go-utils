package text_test

import (
	"fmt"

	"github.com/shouni/go-utils/text"
)

func ExampleCleanStringFromEmojis() {
	cleaned := text.CleanStringFromEmojis("Hello 👋   World 🌍!")
	fmt.Println(cleaned)
	// Output: Hello World !
}

func ExampleNormalizeText() {
	normalized := text.NormalizeText("  too\n\tmany   spaces  ")
	fmt.Println(normalized)
	// Output: too many spaces
}

func ExampleTruncate() {
	truncated := text.Truncate("This is a long sentence.", 7, "...")
	fmt.Println(truncated)
	// Output: This is...
}

func ExampleParseCommaSeparatedList() {
	items := text.ParseCommaSeparatedList("go, rust ,, python")
	fmt.Println(items)
	// Output: [go rust python]
}
