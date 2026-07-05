package timeutil_test

import (
	"fmt"
	"time"

	"github.com/shouni/go-utils/timeutil"
)

func ExampleFormatJST() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(timeutil.FormatJST(utcTime, "2006-01-02 15:04:05"))
	// Output: 2025-01-01 09:00:00
}

func ExampleToJST() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	jstTime := timeutil.ToJST(utcTime)
	fmt.Println(jstTime.Hour())
	// Output: 9
}

func ExampleFormatJSTString() {
	formatted, err := timeutil.FormatJSTString("18:30", "15:04", "15時04分")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(formatted)
	// Output: 18時30分
}
