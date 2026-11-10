package strlist_test

import (
	"fmt"

	"github.com/shouni/go-utils/strlist"
)

func ExampleNormalize() {
	// Pass a value the configuration library has merely split on commas.
	items := strlist.Normalize([]string{"go", " rust ", "", "go", "python"})
	fmt.Println(items)
	// Output: [go rust python]
}

func ExampleNormalizeFold() {
	// Tidy a case-insensitive setting such as hostnames.
	hosts := strlist.NormalizeFold([]string{"Example.com", " EXAMPLE.com ", "api.example.com"})
	fmt.Println(hosts)
	// Output: [example.com api.example.com]
}
