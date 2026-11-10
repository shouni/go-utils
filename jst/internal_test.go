package jst

import (
	"testing"
	"time"
)

// TestLoadLocationOrFallback checks both branches: the location loads, or it does not.
// Location() is cached with sync.OnceValue, so the fallback branch cannot be reached from
// outside; this internal test exercises loadLocationOrFallback directly.
func TestLoadLocationOrFallback(t *testing.T) {
	tests := []struct {
		name     string
		zone     string
		wantName string
	}{
		{"valid zone name loads as is", locationName, locationName},
		{"invalid zone name falls back to FixedZone", "Not/A/Real/Zone", "JST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := loadLocationOrFallback(tt.zone)

			if loc.String() != tt.wantName {
				t.Errorf("loadLocationOrFallback(%q).String() = %q, want %q", tt.zone, loc.String(), tt.wantName)
			}

			// Either path yields UTC+9.
			if _, offset := time.Now().In(loc).Zone(); offset != 9*60*60 {
				t.Errorf("loadLocationOrFallback(%q) offset = %d, want %d", tt.zone, offset, 9*60*60)
			}
		})
	}
}
