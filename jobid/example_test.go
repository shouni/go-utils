package jobid_test

import (
	"fmt"
	"sort"
	"time"

	"github.com/shouni/go-utils/jobid"
)

func ExampleValidate() {
	fmt.Println(jobid.Validate("20260725123456-abcd1234"))
	fmt.Println(jobid.Validate("../etc/passwd"))
	// Output:
	// <nil>
	// invalid job id: "../etc/passwd"
}

func ExampleSanitize() {
	// Drop path elements that slipped into external input, then validate.
	id, err := jobid.Sanitize("../../20260725123456-abcd1234")
	fmt.Println(id, err)
	// Output: 20260725123456-abcd1234 <nil>
}

func ExampleIsValid() {
	fmt.Println(jobid.IsValid("video-recipe-20260725-150405-a1b2c3d4"))
	fmt.Println(jobid.IsValid("-leading-hyphen"))
	// Output:
	// true
	// false
}

func ExampleCreatedAt() {
	// IDs are minted in UTC, so the result is UTC as well.
	t, err := jobid.CreatedAt("video-recipe-20260725-150405-a1b2c3d4")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(t.Format(time.RFC3339))
	// Output: 2026-07-25T15:04:05Z
}

func ExampleCreatedAt_legacyFormats() {
	// IDs minted before New existed can be read too.
	for _, id := range []string{
		"c20260725-150405-1a2b3c4d",
		"20260725150405-a1b2c3d4",
	} {
		t, _ := jobid.CreatedAt(id)
		fmt.Println(t.Format(time.RFC3339))
	}
	// Output:
	// 2026-07-25T15:04:05Z
	// 2026-07-25T15:04:05Z
}

func ExampleSortKey() {
	// Sort IDs with different purpose prefixes newest-first by creation time.
	ids := []string{"regen-zip-20260725-150405-aa", "recipe-20260726-090000-bb"}
	sort.Slice(ids, func(i, j int) bool {
		return jobid.SortKey(ids[i]) > jobid.SortKey(ids[j])
	})
	fmt.Println(ids)
	// Output: [recipe-20260726-090000-bb regen-zip-20260725-150405-aa]
}
