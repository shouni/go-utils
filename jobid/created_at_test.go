package jobid_test

import (
	"errors"
	"testing"
	"time"

	"github.com/shouni/go-utils/jobid"
)

// want is 2026-08-03 02:41:06 UTC (11:41 the same day in JST).
var want = time.Date(2026, time.August, 3, 2, 41, 6, 0, time.UTC)

func TestCreatedAt(t *testing.T) {
	tests := []struct {
		name  string
		jobID string
	}{
		{"New が生成する形式", "recipe-20260803-024106-a1b2c3d4e5f6"},
		{"プレフィックスにハイフンを含む形式", "video-recipe-20260803-024106-a1b2c3d4e5f6"},
		{"プレフィックスが日付に直結する形式", "c20260803-024106-1a2b3c4d"},
		{"日付と時刻が分割されない形式", "20260803024106-a1b2c3d4"},
		{"乱数部が数字のみでも時刻部を優先する", "recipe-20260803-024106-123456789012"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jobid.CreatedAt(tt.jobID)
			if err != nil {
				t.Fatalf("CreatedAt(%q) returned an unexpected error: %v", tt.jobID, err)
			}
			if !got.Equal(want) {
				t.Errorf("CreatedAt(%q) = %v, want %v", tt.jobID, got, want)
			}
			// The generator mints in UTC, so the result is UTC too.
			if got.Location() != time.UTC {
				t.Errorf("CreatedAt(%q).Location() = %v, want UTC", tt.jobID, got.Location())
			}
		})
	}
}

func TestCreatedAt_NoTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		jobID string
	}{
		{"空文字", ""},
		{"時刻を含まない", "recipe-abcdef"},
		{"桁数が足りない", "recipe-2026080-024106-abcd"},
		{"月日が範囲外", "recipe-20261345-024106-abcd"},
		{"時刻が範囲外", "recipe-20260803-254106-abcd"},
		{"日付として妥当でも下限より前なら乱数の偶然とみなす", "recipe-19990101-000000-abcd"},
		{"分割されない形式でも下限を適用する", "19991231235959-abcd"},
		{"ハイフンなしの乱数のみ", "abcdefghijklmn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jobid.CreatedAt(tt.jobID)
			if err == nil {
				t.Fatalf("CreatedAt(%q) returned no error", tt.jobID)
			}
			if !errors.Is(err, jobid.ErrNoTimestamp) {
				t.Errorf("CreatedAt(%q) error does not wrap ErrNoTimestamp: %v", tt.jobID, err)
			}
		})
	}
}

// TestCreatedAt_RoundTrip checks that the creation time can be recovered from an ID New generated.
// It fails when the generated format changes.
func TestCreatedAt_RoundTrip(t *testing.T) {
	for _, prefix := range []string{"", "job", "video-recipe", "regen-keyframe"} {
		before := time.Now().UTC().Truncate(time.Second)
		id, err := jobid.New(prefix)
		if err != nil {
			t.Fatalf("New(%q) failed: %v", prefix, err)
		}
		after := time.Now().UTC()

		got, err := jobid.CreatedAt(id)
		if err != nil {
			t.Fatalf("CreatedAt(%q) failed: %v", id, err)
		}
		if got.Before(before) || got.After(after) {
			t.Errorf("CreatedAt(%q) = %v, want between %v and %v", id, got, before, after)
		}
	}
}

func TestSortKey(t *testing.T) {
	// IDs of different formats mixed together still sort newest-first by creation time.
	ids := []string{
		"recipe-20260803-024106-a1b2c3d4e5f6", // 2026-08-03 02:41:06
		"20260801120000-a1b2c3d4",             // 2026-08-01 12:00:00
		"c20260805-093000-1a2b3c4d",           // 2026-08-05 09:30:00
		"broken-id-without-timestamp",         // no timestamp
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = jobid.SortKey(id)
	}

	wantKeys := []string{"20260803024106", "20260801120000", "20260805093000", ""}
	for i := range ids {
		if keys[i] != wantKeys[i] {
			t.Errorf("SortKey(%q) = %q, want %q", ids[i], keys[i], wantKeys[i])
		}
	}

	// Descending key order is newest-first, with the timestamp-less ID last.
	descending := []string{keys[2], keys[0], keys[1], keys[3]}
	for i := 0; i+1 < len(descending); i++ {
		if descending[i] <= descending[i+1] {
			t.Errorf("descending SortKey order is not newest-first by creation time: %q", descending)
			break
		}
	}
}
