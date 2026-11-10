package jobid

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	valid := []string{
		"20260725123456-abcd1234",               // legacy unprefixed format
		"video-recipe-20260725-150405-a1b2c3d4", // prefix containing a hyphen
		"job_with_underscores",
		"a",
		strings.Repeat("a", MaxLength),
	}
	for _, id := range valid {
		if err := Validate(id); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", id, err)
		}
	}

	// Check the expected error too, so callers can branch on the kind of error.
	invalid := []struct {
		reason string
		id     string
		want   error
	}{
		{"empty", "", ErrEmpty},
		{"leading hyphen", "-leading-hyphen", ErrInvalidFormat},
		{"leading underscore", "_leading_underscore", ErrInvalidFormat},
		{"contains a path separator", "has/slash", ErrInvalidFormat},
		{"parent directory reference", "..", ErrInvalidFormat},
		{"path traversal", "../etc/passwd", ErrInvalidFormat},
		{"contains a space", "has space", ErrInvalidFormat},
		{"contains a dot", "has.dot", ErrInvalidFormat},
		{"non-ASCII", "日本語", ErrInvalidFormat},
		{"over the length limit", strings.Repeat("a", MaxLength+1), ErrTooLong},
	}
	for _, tt := range invalid {
		if err := Validate(tt.id); !errors.Is(err, tt.want) {
			t.Errorf("Validate(%q) = %v, want %v (%s)", tt.id, err, tt.want, tt.reason)
		}
	}
}

func TestSanitize(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"20260725123456-abcd1234":        "20260725123456-abcd1234",
		"  20260725123456-abcd1234  ":    "20260725123456-abcd1234",
		"prefix/20260725123456-abcd1234": "20260725123456-abcd1234",
		"../../20260725123456-abcd1234":  "20260725123456-abcd1234",
	}
	for input, want := range tests {
		got, err := Sanitize(input)
		if err != nil {
			t.Errorf("Sanitize(%q) error = %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", input, got, want)
		}
	}
}

// Input whose last element is still malformed is rejected.
func TestSanitizeRejectsInvalidBase(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"../../etc/pass wd", "", "dir/-leading", "dir/日本語", "/"} {
		if got, err := Sanitize(input); err == nil {
			t.Errorf("Sanitize(%q) = %q, want an error", input, got)
		}
	}
}

// Sanitize keeps only the last element, so a trailing slash is ignored
// (path.Base behaviour). Directory-like input passes if its last element is valid.
func TestSanitizeUsesLastPathElement(t *testing.T) {
	t.Parallel()

	got, err := Sanitize("prefix/20260725123456-abcd1234/")
	if err != nil {
		t.Fatalf("Sanitize() error = %v", err)
	}
	if got != "20260725123456-abcd1234" {
		t.Fatalf("Sanitize() = %q", got)
	}
}

func TestNewProducesValidID(t *testing.T) {
	t.Parallel()

	id, err := New("video-recipe")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := Validate(id); err != nil {
		t.Fatalf("New() produced an invalid id %q: %v", id, err)
	}
	if !strings.HasPrefix(id, "video-recipe-") {
		t.Fatalf("New() = %q, want the prefix preserved", id)
	}
}

func TestNewIsUnique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{}, 100)
	for range 100 {
		id, err := New("job")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("New() returned a duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

// The creation time comes first, so lexicographic order matches newest-first order.
// This guarantees that a key-ordered storage listing can serve as the history order as is.
func TestNewIsLexicographicallySortable(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 7, 25, 12, 34, 56, 0, time.UTC)
	older, err := newAt("job", base)
	if err != nil {
		t.Fatalf("newAt() error = %v", err)
	}
	newer, err := newAt("job", base.Add(time.Second))
	if err != nil {
		t.Fatalf("newAt() error = %v", err)
	}

	if older >= newer {
		t.Fatalf("expected %q < %q", older, newer)
	}
}

// Issuing an ID never fails over the spelling of the prefix.
func TestNewNormalizesPrefix(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"":               "job-",
		"   ":            "job-",
		"--":             "job-",
		"Regen Keyframe": "regenkeyframe-",
		"_private":       "private-",
		"日本語ジョブ":         "job-",
	}
	for prefix, wantPrefix := range tests {
		id, err := New(prefix)
		if err != nil {
			t.Errorf("New(%q) error = %v", prefix, err)
			continue
		}
		if !strings.HasPrefix(id, wantPrefix) {
			t.Errorf("New(%q) = %q, want prefix %q", prefix, id, wantPrefix)
		}
		if err := Validate(id); err != nil {
			t.Errorf("New(%q) produced an invalid id %q: %v", prefix, id, err)
		}
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	if !IsValid("20260725123456-abcd1234") {
		t.Error("IsValid() = false, want true")
	}
	if IsValid("../etc/passwd") {
		t.Error("IsValid() = true, want false")
	}
}

// Empty input yields an error describing the input itself, not the "." that path.Base returns.
func TestSanitizeEmptyReportsMissingInput(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "   "} {
		_, err := Sanitize(input)
		if !errors.Is(err, ErrEmpty) {
			t.Errorf("Sanitize(%q) = %v, want %v", input, err, ErrEmpty)
		}
		if err != nil && strings.Contains(err.Error(), ".") {
			t.Errorf("Sanitize(%q) error mentions a path element that was not in the input: %v", input, err)
		}
	}
}

// A prefix that is too long is truncated rather than failing ID issuance.
func TestNewTruncatesLongPrefix(t *testing.T) {
	t.Parallel()

	id, err := New(strings.Repeat("a", MaxLength*2))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := Validate(id); err != nil {
		t.Fatalf("New() produced an invalid id %q: %v", id, err)
	}
	if len(id) != MaxLength {
		t.Fatalf("len(New()) = %d, want %d (the prefix should fill the id up to the limit)", len(id), MaxLength)
	}
}

// New adds the separating hyphen, so a trailing separator on the prefix is not doubled.
func TestNewTrimsTrailingSeparatorInPrefix(t *testing.T) {
	t.Parallel()

	for _, prefix := range []string{"video-recipe-", "video-recipe__"} {
		id, err := New(prefix)
		if err != nil {
			t.Errorf("New(%q) error = %v", prefix, err)
			continue
		}
		if !strings.HasPrefix(id, "video-recipe-2") {
			t.Errorf("New(%q) = %q, want the separator collapsed", prefix, id)
		}
	}
}
