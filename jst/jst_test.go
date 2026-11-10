package jst_test

import (
	"errors"
	"testing"
	"time"

	"github.com/shouni/go-utils/jst"
)

// jstOffset is the offset between UTC and JST (9 hours).
const jstOffset = 9 * time.Hour

// isJST reports whether a location name means JST ("Asia/Tokyo", or the fallback "JST").
func isJST(name string) bool {
	return name == "Asia/Tokyo" || name == "JST"
}

// TestLocation checks that Location() returns the right location and that it is cached.
func TestLocation(t *testing.T) {
	loc1 := jst.Location()
	if !isJST(loc1.String()) {
		t.Errorf("Location().String() = %q, want \"Asia/Tokyo\" or \"JST\"", loc1.String())
	}

	// If the sync.OnceValue cache works, the pointers are identical.
	if loc2 := jst.Location(); loc1 != loc2 {
		t.Error("Location() returned a different pointer on each call; it is not cached")
	}

	if _, offset := time.Now().In(loc1).Zone(); offset != int(jstOffset.Seconds()) {
		t.Errorf("Location() offset = %d, want %d", offset, int(jstOffset.Seconds()))
	}
}

// TestNow checks that the current time is returned with JST (UTC+9) zone information.
func TestNow(t *testing.T) {
	now := jst.Now()

	if !isJST(now.Location().String()) {
		t.Errorf("Now().Location() = %q, want \"Asia/Tokyo\" or \"JST\"", now.Location().String())
	}

	if _, offset := now.Zone(); offset != int(jstOffset.Seconds()) {
		t.Errorf("Now() offset = %d, want %d", offset, int(jstOffset.Seconds()))
	}

	// The absolute time is close to now (a 1-second tolerance allows for execution speed).
	if diff := time.Since(now); diff > time.Second || diff < -time.Second {
		t.Errorf("Now() is off from the current absolute time, diff: %v", diff)
	}
}

// TestFrom checks that a UTC time is converted to JST correctly.
func TestFrom(t *testing.T) {
	utc := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	want := time.Date(2025, time.January, 1, 9, 0, 0, 0, jst.Location())

	got := jst.From(utc)

	// The absolute time is unchanged.
	if !got.Equal(want) {
		t.Errorf("From(%v) = %v, want %v", utc, got, want)
	}
	if got.Location() != jst.Location() {
		t.Errorf("From(%v).Location() = %q, want %q", utc, got.Location(), jst.Location())
	}
}

// TestFormat checks that the time is converted to JST and then formatted with the given layout.
func TestFormat(t *testing.T) {
	utc := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		layout string
		want   string
	}{
		{"任意のレイアウト", "2006/01/02 15:04", "2025/01/01 09:00"},
		{"LayoutDisplay", jst.LayoutDisplay, "2025-01-01 09:00 JST"},
		{"LayoutTimestamp", jst.LayoutTimestamp, "2025/01/01 09:00:00 JST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jst.Format(utc, tt.layout); got != tt.want {
				t.Errorf("Format(%v, %q) = %q, want %q", utc, tt.layout, got, tt.want)
			}
		})
	}
}

// TestParse checks that a string without time zone information is interpreted as JST.
func TestParse(t *testing.T) {
	got, err := jst.Parse("2025-01-01 18:30", "2006-01-02 15:04")
	if err != nil {
		t.Fatalf("Parse() returned an unexpected error: %v", err)
	}

	want := time.Date(2025, time.January, 1, 18, 30, 0, 0, jst.Location())
	if !got.Equal(want) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}

	// It is always interpreted as UTC+9, independent of the host's time.Local.
	if _, offset := got.Zone(); offset != int(jstOffset.Seconds()) {
		t.Errorf("Parse() offset = %d, want %d", offset, int(jstOffset.Seconds()))
	}
}

// TestParse_Invalid checks that invalid input returns an error that wraps the original error.
func TestParse_Invalid(t *testing.T) {
	_, err := jst.Parse("25:00", "15:04")
	if err == nil {
		t.Fatal("Parse() returned no error for invalid input")
	}

	// It is wrapped with %w, so the time.ParseError can be extracted.
	if _, ok := errors.AsType[*time.ParseError](err); !ok {
		t.Errorf("Parse() error does not wrap *time.ParseError: %v", err)
	}
}

// TestFormatDisplayAndTimestamp checks that each delegates to its layout constant.
// The output of each constant itself is checked by TestFormat.
func TestFormatDisplayAndTimestamp(t *testing.T) {
	utcTime := time.Date(2026, time.July, 25, 6, 4, 5, 0, time.UTC)

	if got, want := jst.FormatDisplay(utcTime), jst.Format(utcTime, jst.LayoutDisplay); got != want {
		t.Errorf("FormatDisplay() = %q, want %q", got, want)
	}
	if got, want := jst.FormatTimestamp(utcTime), jst.Format(utcTime, jst.LayoutTimestamp); got != want {
		t.Errorf("FormatTimestamp() = %q, want %q", got, want)
	}
}
