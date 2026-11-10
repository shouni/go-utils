// Package jst provides utility functions that simplify time handling built around
// Japan Standard Time (JST), such as converting to and formatting in JST.
//
// The package is for the display layer. Keep persisted times in UTC and convert them
// to JST just before they are shown on screen or written to a log.
package jst

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Layouts for display, kept here so the notation cannot diverge between projects.
const (
	// LayoutDisplay is for minute-precision display such as history listings.
	// The trailing MST is a layout directive, so it is replaced by the actual zone abbreviation (JST).
	LayoutDisplay = "2006-01-02 15:04 MST"

	// LayoutTimestamp is for second-precision display such as notification footers.
	// The trailing JST is a literal and is always printed as "JST".
	LayoutTimestamp = "2006/01/02 15:04:05 JST"
)

const locationName = "Asia/Tokyo"

// The cache behind Location. It is not an eagerly initialised package variable because
// that would emit the load-failure warning before the consumer has installed its slog handler.
var locationOnce = sync.OnceValue(func() *time.Location {
	return loadLocationOrFallback(locationName)
})

// Location loads the "Asia/Tokyo" time.Location once and returns a pointer to it.
// Later calls return the cached value, so no call touches the filesystem again.
func Location() *time.Location {
	return locationOnce()
}

// loadLocationOrFallback loads the Location for the given IANA time zone name.
// If loading fails, it logs a warning and falls back to a FixedZone at UTC+9.
func loadLocationOrFallback(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		slog.Warn(
			"failed to load location, falling back to FixedZone",
			slog.String("location", name),
			slog.String("fallback", "FixedZone (UTC+9)"),
			slog.Any("error", err),
		)
		return time.FixedZone("JST", 9*60*60)
	}
	return loc
}

// Now returns the current time in Japan Standard Time (JST).
func Now() time.Time {
	return time.Now().In(Location())
}

// From converts the given time.Time to JST.
func From(t time.Time) time.Time {
	return t.In(Location())
}

// Format converts the given time.Time to JST and then formats it with the given layout.
func Format(t time.Time, layout string) string {
	return From(t).Format(layout)
}

// FormatDisplay formats with LayoutDisplay. Use it to keep times shown side by side,
// such as in history listings, in one notation.
//
// Passing the layout constant to Format gives the same result, but the two constants
// treat their endings differently, so mixing them up is easy to miss; this lets callers
// choose by name.
func FormatDisplay(t time.Time) string {
	return Format(t, LayoutDisplay)
}

// FormatTimestamp formats with LayoutTimestamp. Use it where the output should state
// explicitly that it is always JST, such as notification footers.
func FormatTimestamp(t time.Time) string {
	return Format(t, LayoutTimestamp)
}

// Parse interprets a time string without time zone information as JST and returns a time.Time.
//
// time.Parse matches a zone abbreviation against the host's time.Local, so the offset
// shifts with the runtime's time zone setting. Parse passes Location explicitly to remove
// that dependence on the environment.
func Parse(value, layout string) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q as %q: %w", value, layout, err)
	}
	return t, nil
}
