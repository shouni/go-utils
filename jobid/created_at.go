package jobid

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNoTimestamp reports that no creation time could be extracted from a job ID.
// Classify it with errors.Is.
var ErrNoTimestamp = errors.New("job id has no embedded timestamp")

// timestampLayout is the format of the time embedded in a job ID.
const timestampLayout = "20060102150405"

// minTimestamp is the earliest time accepted as an embedded timestamp.
//
// A job ID contains digits from random data, which can happen to form a run of 14 digits.
// Range checks on month, day, hour, minute and second alone cannot reject a value such as
// "0241-06-12 03:45:06", so times at which no job could have existed are excluded.
var minTimestamp = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

// CreatedAt returns the creation time embedded in a job ID, in UTC.
//
// The generator (New) mints IDs in UTC, so the result is UTC as well. To show it on
// screen, convert it to JST with go-utils/jst.
//
// Besides the format New generates, it also reads the formats each service minted on its
// own before New existed: job IDs are used in the paths of artifacts in object storage,
// so IDs minted in the past live on. The three supported formats are:
//
//	{prefix}-20060102-150405-{random}   the format New generates (prefix may contain hyphens)
//	c20060102-150405-{random}           a prefix attached directly to the date
//	20060102150405-{random}             date and time not separated
//
// When no time can be extracted, it returns an error wrapping ErrNoTimestamp.
// It does not validate the ID; combine it with Validate if needed.
func CreatedAt(jobID string) (time.Time, error) {
	parts := strings.Split(jobID, "-")

	// Scan from the front and adopt the first time that parses as valid.
	for i, part := range parts {
		// Date and time not separated.
		if t, ok := parseTimestamp(trimLeadingNonDigits(part)); ok {
			return t, nil
		}

		// Date (8 digits) and time (6 digits) separated by a hyphen.
		// Leading non-digits are dropped to absorb a prefix attached directly, as in "c20260803".
		if i+1 < len(parts) {
			if date := trimLeadingNonDigits(part); len(date) == 8 && len(parts[i+1]) == 6 {
				if t, ok := parseTimestamp(date + parts[i+1]); ok {
					return t, nil
				}
			}
		}
	}

	return time.Time{}, fmt.Errorf("%w: %q", ErrNoTimestamp, jobID)
}

// SortKey returns a key for sorting job IDs by creation time.
// It returns an empty string for an ID with no extractable time.
//
// Comparing IDs as plain strings sorts by prefix when IDs with different per-purpose
// prefixes share one list (they differ before the time part is reached).
func SortKey(jobID string) string {
	t, err := CreatedAt(jobID)
	if err != nil {
		return ""
	}
	return t.Format(timestampLayout)
}

// parseTimestamp interprets a 14-digit string as a time in UTC.
func parseTimestamp(value string) (time.Time, bool) {
	if len(value) != len(timestampLayout) || !isDigits(value) {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(timestampLayout, value, time.UTC)
	if err != nil || t.Before(minTimestamp) {
		return time.Time{}, false
	}
	return t, true
}

// trimLeadingNonDigits removes any non-digit characters at the start.
func trimLeadingNonDigits(value string) string {
	return strings.TrimLeftFunc(value, func(r rune) bool {
		return r < '0' || r > '9'
	})
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
