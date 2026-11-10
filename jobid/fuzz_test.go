package jobid

import (
	"strings"
	"testing"
	"time"
)

// FuzzSanitize checks that whatever Sanitize returns always passes Validate.
//
// Validation is a security boundary. Job IDs appear in URL paths and storage paths, and
// Sanitize is the last gate before a value from outside is built into them. If it ever lets
// "invalid but accepted" through, path separators and parent directory references flow
// straight downstream. A table test only tries the shapes someone thought of, so fuzzing
// fills the rest.
func FuzzSanitize(f *testing.F) {
	seeds := []string{
		"20260725123456-abcd1234",
		"music-20260725-123456-abcdef123456",
		"../../etc/passwd",
		"a/b/c",
		"  spaced  ",
		"-leading-hyphen",
		"_leading-underscore",
		".",
		"..",
		"/",
		"",
		"あいうえお",
		strings.Repeat("a", MaxLength),
		strings.Repeat("a", MaxLength+1),
		"a\x00b",
		"a%2Fb",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		safe, err := Sanitize(raw)
		if err != nil {
			if safe != "" {
				t.Fatalf("Sanitize(%q) failed but returned %q", raw, safe)
			}
			return
		}

		// On success the value always passes Validate.
		if err := Validate(safe); err != nil {
			t.Fatalf("Sanitize(%q) = %q, which Validate rejects: %v", raw, safe, err)
		}

		// No dangerous shape remains. This follows from Validate's character set,
		// but it is the boundary itself, so check it directly as well.
		if strings.ContainsAny(safe, `/\.%`+"\x00") || strings.TrimSpace(safe) != safe {
			t.Fatalf("Sanitize(%q) = %q, which is not path-safe", raw, safe)
		}
		if len(safe) > MaxLength {
			t.Fatalf("Sanitize(%q) = %q (%d bytes), over MaxLength", raw, safe, len(safe))
		}

		// Idempotent: passing an accepted value through again leaves it unchanged.
		again, err := Sanitize(safe)
		if err != nil || again != safe {
			t.Fatalf("Sanitize is not idempotent: %q -> %q -> (%q, %v)", raw, safe, again, err)
		}
	})
}

// FuzzCreatedAt checks that a time CreatedAt returns on success is plausible and agrees
// with SortKey.
//
// What it parses is "the three formats minted in the past that still remain in storage",
// using a hand-written scan over the hyphen-separated elements from the front. Listing order
// is decided by this result, so a misparse silently breaks the history order.
func FuzzCreatedAt(f *testing.F) {
	seeds := []string{
		"20260725123456-abcd1234",
		"c20260803-101112-abcd1234",
		"music-20260725-123456-abcdef123456",
		"video-gen-20260725-123456-abcdef123456",
		"no-timestamp-here",
		"19990101000000-old",
		"20261332999999-invalid",
		"-------",
		"",
		"12345678-901234",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		got, err := CreatedAt(raw)
		key := SortKey(raw)

		if err != nil {
			if key != "" {
				t.Fatalf("CreatedAt(%q) failed but SortKey returned %q", raw, key)
			}
			return
		}

		// A time before minting began can only come from a misparse.
		if got.Before(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("CreatedAt(%q) = %v, before the minimum", raw, got)
		}
		if got.Location() != time.UTC {
			t.Fatalf("CreatedAt(%q) = %v, want UTC", raw, got)
		}
		// The sort key is derived from the same time.
		if want := got.Format(timestampLayout); key != want {
			t.Fatalf("SortKey(%q) = %q, want %q", raw, key, want)
		}
	})
}
