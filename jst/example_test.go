package jst_test

import (
	"fmt"
	"time"

	"github.com/shouni/go-utils/jst"
)

func ExampleFormat() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(jst.Format(utcTime, "2006-01-02 15:04:05"))
	// Output: 2025-01-01 09:00:00
}

func ExampleFormat_layoutDisplay() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(jst.Format(utcTime, jst.LayoutDisplay))
	// Output: 2025-01-01 09:00 JST
}

func ExampleFrom() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(jst.From(utcTime).Hour())
	// Output: 9
}

func ExampleParse() {
	// タイムゾーン情報を含まない文字列を、ホストの設定によらず JST として解釈する。
	t, err := jst.Parse("18:30", "15:04")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(jst.Format(t, "15時04分"))
	// Output: 18時30分
}

func ExampleFormatDisplay() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(jst.FormatDisplay(utcTime))
	// Output: 2025-01-01 09:00 JST
}

func ExampleFormatTimestamp() {
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(jst.FormatTimestamp(utcTime))
	// Output: 2025/01/01 09:00:00 JST
}
