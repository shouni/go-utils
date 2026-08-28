package jst

import (
	"testing"
	"time"
)

// TestLoadLocationOrFallback は、Location のロード可否による分岐を確認します。
// Location() は sync.OnceValue でキャッシュされフォールバック分岐を外部から到達させられないため、
// この内部テストで loadLocationOrFallback を直接検証しています。
func TestLoadLocationOrFallback(t *testing.T) {
	tests := []struct {
		name     string
		zone     string
		wantName string
	}{
		{"有効なタイムゾーン名はそのままロードされる", locationName, locationName},
		{"不正なタイムゾーン名は FixedZone にフォールバックする", "Not/A/Real/Zone", "JST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := loadLocationOrFallback(tt.zone)

			if loc.String() != tt.wantName {
				t.Errorf("loadLocationOrFallback(%q).String() = %q, want %q", tt.zone, loc.String(), tt.wantName)
			}

			// いずれの経路でも UTC+9 であること。
			if _, offset := time.Now().In(loc).Zone(); offset != 9*60*60 {
				t.Errorf("loadLocationOrFallback(%q) のオフセット = %d, want %d", tt.zone, offset, 9*60*60)
			}
		})
	}
}
