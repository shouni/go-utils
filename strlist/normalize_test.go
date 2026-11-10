package strlist_test

import (
	"reflect"
	"testing"

	"github.com/shouni/go-utils/strlist"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "clean input unchanged",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			// Splitting in the configuration library does not trim whitespace.
			name:     "trims surrounding spaces",
			input:    []string{"  apple  ", "banana", "  cherry "},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			// These come from a trailing comma or consecutive commas.
			name:     "drops empty elements",
			input:    []string{"", "apple", "", "banana", ""},
			expected: []string{"apple", "banana"},
		},
		{
			name:     "drops whitespace-only elements",
			input:    []string{" ", "apple", "   "},
			expected: []string{"apple"},
		},
		{
			// Elements that become equal after trimming count as duplicates too.
			name:     "drops duplicates keeping first order",
			input:    []string{"banana", "apple", " banana ", "cherry", "apple"},
			expected: []string{"banana", "apple", "cherry"},
		},
		{
			name:     "empty slice stays empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "nil becomes empty slice",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "all blank becomes empty slice",
			input:    []string{" ", "", "  "},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := strlist.Normalize(tt.input)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Normalize(%q) = %v, want %v", tt.input, actual, tt.expected)
			}
		})
	}
}

// The input slice is not modified. The caller may keep the original values, for example to
// log the raw setting.
func TestNormalizeDoesNotMutateInput(t *testing.T) {
	input := []string{" a ", "", "a", "b"}
	want := []string{" a ", "", "a", "b"}

	strlist.Normalize(input)

	if !reflect.DeepEqual(input, want) {
		t.Errorf("input was modified: %v, want %v", input, want)
	}
}

// Case differences count as duplicates, and the result is lowercased.
func TestNormalizeFold(t *testing.T) {
	got := strlist.NormalizeFold([]string{"Example.com", " EXAMPLE.com ", "", "Other.example"})
	want := []string{"example.com", "other.example"}

	if len(got) != len(want) {
		t.Fatalf("NormalizeFold() = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("NormalizeFold() = %q, want %q", got, want)
		}
	}
}

// Normalize stays case-sensitive (the difference from NormalizeFold).
func TestNormalizeKeepsCase(t *testing.T) {
	got := strlist.Normalize([]string{"Example.com", "example.com"})
	if len(got) != 2 {
		t.Fatalf("Normalize() = %q, want both entries kept", got)
	}
}
