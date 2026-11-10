// Package jobid generates, validates and normalises identifiers for asynchronous jobs (job IDs).
//
// A job ID appears both as an HTTP route parameter and as a path element in object
// storage, so validation doubles as a security boundary. When several services exchange
// the same job IDs and disagree on "what counts as a valid ID", one rejects IDs the
// other issued, or only one of them lets path traversal through.
// This package keeps that boundary in one place.
package jobid

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
)

// MaxLength is the maximum length of a job ID.
// Only ASCII characters are allowed, so the character count equals the byte count.
const MaxLength = 128

// Errors returned by Validate and Sanitize. Classify them with errors.Is.
//
// An HTTP handler turns each of them into a 400, but the kinds are kept separate
// because response messages and metric labels may need to differ.
var (
	// ErrEmpty reports that the job ID is empty (or whitespace only).
	ErrEmpty = errors.New("job id is required")

	// ErrTooLong reports that the job ID exceeds MaxLength.
	ErrTooLong = errors.New("job id is too long")

	// ErrInvalidFormat reports that the job ID contains a disallowed character.
	ErrInvalidFormat = errors.New("invalid job id")
)

// Validate checks that a job ID is in a form that is safe in routes and storage paths.
// A valid ID starts with an alphanumeric character, continues with alphanumerics,
// hyphens and underscores only, and is at most MaxLength characters long.
//
// The first character is restricted to alphanumerics so that a value starting with `-`
// or `_` is not interpreted unexpectedly as a command-line flag or URL query syntax.
// Narrowing the character set structurally excludes the path separator (`/`), parent
// directory references (`..`) and URL-encoded characters.
func Validate(jobID string) error {
	if jobID == "" {
		return ErrEmpty
	}
	if len(jobID) > MaxLength {
		return fmt.Errorf("%w: %d characters (max %d)", ErrTooLong, len(jobID), MaxLength)
	}

	// Every allowed character is ASCII, so the ID can be scanned byte by byte.
	// A multi-byte character is rejected because its leading byte is outside the allowed set.
	if !isAlphanumeric(jobID[0]) {
		return fmt.Errorf("%w: %q", ErrInvalidFormat, jobID)
	}
	for i := 1; i < len(jobID); i++ {
		if !isAllowedInID(jobID[i]) {
			return fmt.Errorf("%w: %q", ErrInvalidFormat, jobID)
		}
	}
	return nil
}

// IsValid reports whether Validate returns no error.
func IsValid(jobID string) bool {
	return Validate(jobID) == nil
}

// Sanitize normalises a value that may be path-shaped into a safe job ID.
//
// It is meant for values received from outside, before they are built into a storage
// path. Extra path elements added by upstream routing or normalisation are dropped before
// Validate runs, so an input such as `../../etc/passwd` is cut down to `passwd`, and an
// invalid format is rejected.
func Sanitize(jobID string) (string, error) {
	trimmed := strings.TrimSpace(jobID)

	// path.Base returns "." for an empty string, so reject it first. A "." that was
	// never in the input showing up in the error message misleads whoever reads it.
	if trimmed == "" {
		return "", ErrEmpty
	}

	safe := path.Base(trimmed)
	if err := Validate(safe); err != nil {
		return "", err
	}
	return safe, nil
}

// randomBytes is the number of bytes in the random tail of an ID (twice as many hex digits).
const randomBytes = 6

// generatedSuffixLength is the length of the part New appends after the prefix.
const generatedSuffixLength = len("-20060102-150405-") + 2*randomBytes

// maxPrefixLength is the longest prefix for which the generated ID stays within MaxLength.
const maxPrefixLength = MaxLength - generatedSuffixLength

// New generates a job ID with the given prefix.
//
// The format is `{prefix}-{yyyymmdd}-{hhmmss}-{12 random hex digits}`, with the time in UTC.
// An empty prefix becomes "job".
//
// Lexicographic order equals newest-first only when every ID in the list shares the same
// prefix. Object storage lists keys in order, so under that condition no separate index is
// needed. For a list with mixed prefixes, sort by SortKey.
func New(prefix string) (string, error) {
	return newAt(prefix, time.Now().UTC())
}

// newAt is New with a caller-supplied creation time (for tests).
func newAt(prefix string, now time.Time) (string, error) {
	normalized := normalizePrefix(prefix)

	buf := make([]byte, randomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("job id entropy: %w", err)
	}

	id := fmt.Sprintf("%s-%s-%s", normalized, now.Format("20060102-150405"), hex.EncodeToString(buf))
	if err := Validate(id); err != nil {
		return "", err
	}
	return id, nil
}

// normalizePrefix shapes the prefix so that the generated ID passes Validate.
//
// The prefix is decoration the caller adds to say what the job is for; uniqueness comes
// from the timestamp and the random part. So disallowed characters or excess length are
// not an error: they are dropped, or the prefix is truncated to maxPrefixLength, and
// generation continues (failing to issue an ID over a prefix's spelling is the worse outcome).
func normalizePrefix(prefix string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(prefix)) {
		if r < 0x80 && isAllowedInID(byte(r)) {
			b.WriteByte(byte(r))
		}
	}

	// A prefix that does not start with an alphanumeric would make the whole ID invalid,
	// so leading separators are removed. Trailing separators are dropped too: New adds the
	// separating hyphen itself, and keeping them would double it as in "prefix--20060102-...".
	normalized := strings.Trim(b.String(), "_-")
	if normalized == "" {
		return "job"
	}

	// Only ASCII remains, so cutting by bytes cannot break a character.
	// The cut may land on a separator, so the tail is trimmed for the same reason.
	if len(normalized) > maxPrefixLength {
		normalized = strings.TrimRight(normalized[:maxPrefixLength], "_-")
	}
	return normalized
}

func isAllowedInID(c byte) bool {
	return isAlphanumeric(c) || c == '-' || c == '_'
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
