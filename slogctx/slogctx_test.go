package slogctx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

// newTestLogger returns a context-aware logger that writes JSON output to buf.
func newTestLogger(buf *bytes.Buffer, level slog.Level) *slog.Logger {
	return slog.New(NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})))
}

// decodeLines decodes the emitted JSON log, one line at a time, into maps.
func decodeLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var entries []map[string]any
	decoder := json.NewDecoder(buf)
	for decoder.More() {
		var entry map[string]any
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode log line: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		raw  string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		// Environment variables often carry surrounding whitespace, so it is trimmed.
		{" warn ", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"", slog.LevelInfo},
		{"nonsense", slog.LevelInfo},
		// slog's relative notation, usable to narrow log volume step by step.
		{"INFO+2", slog.LevelInfo + 2},
		{"error-1", slog.LevelError - 1},
		// A broken numeric part in the relative notation also falls back to the default Info.
		{"INFO+x", slog.LevelInfo},
	}

	for _, tt := range tests {
		if got := ParseLevel(tt.raw); got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.raw, got, tt.want)
		}
	}
}

func TestWithAddsAttrsToEveryRecord(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo)

	ctx := With(context.Background(), slog.String("job_id", "job-1"))
	ctx = With(ctx, slog.String("command", "compose"))

	logger.InfoContext(ctx, "phase started")
	logger.InfoContext(context.Background(), "unrelated")

	entries := decodeLines(t, &buf)
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0]["job_id"] != "job-1" || entries[0]["command"] != "compose" {
		t.Errorf("entry[0] = %v, want it to contain job_id/command", entries[0])
	}
	if _, ok := entries[1]["job_id"]; ok {
		t.Errorf("job_id leaked into a log line from a context without attributes: %v", entries[1])
	}
}

// Work branched from the same context does not pollute each other's attributes.
func TestWithDoesNotLeakBetweenBranches(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo)

	base := With(context.Background(), slog.String("job_id", "job-1"))
	logger.InfoContext(With(base, slog.String("phase", "collect")), "a")
	logger.InfoContext(With(base, slog.String("phase", "publish")), "b")

	entries := decodeLines(t, &buf)
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0]["phase"] != "collect" || entries[1]["phase"] != "publish" {
		t.Errorf("phase = %v / %v, want collect / publish", entries[0]["phase"], entries[1]["phase"])
	}
}

// Context attributes survive re-wrapping through logger.With.
func TestContextAttrsSurviveLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo).With("component", "pipeline")

	logger.InfoContext(With(context.Background(), slog.String("job_id", "job-1")), "msg")

	entries := decodeLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0]["component"] != "pipeline" || entries[0]["job_id"] != "job-1" {
		t.Errorf("entry = %v, want it to contain both component and job_id", entries[0])
	}
}

// Context attributes survive re-wrapping through WithGroup.
func TestContextAttrsSurviveWithGroup(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(NewHandler(base).WithGroup("req"))

	logger.InfoContext(With(context.Background(), slog.String("job_id", "job-1")), "msg")

	entries := decodeLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	group, ok := entries[0]["req"].(map[string]any)
	if !ok {
		t.Fatalf("group req was not emitted: %v", entries[0])
	}
	if group["job_id"] != "job-1" {
		t.Errorf("req.job_id = %v, want job-1", group["job_id"])
	}
}

// countKey counts how many times the key appears in the emitted JSON.
//
// Duplicates vanish once decoded into a map (the last one wins), so the raw output is counted.
// The same sequence appearing in a value would be a false positive, but the tests choose the
// values, so it does not happen.
func countKey(t *testing.T, out []byte, key string) int {
	t.Helper()
	return bytes.Count(out, []byte(`"`+key+`":`))
}

// When the call site passes the same key as the context, the key does not appear twice.
// Two copies are still valid JSON, so nothing fails, and it breaks on the ingesting side
// (see the comment on Handle).
func TestCallerAttrOverridesContextAttr(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo)

	ctx := With(context.Background(), slog.String("job_id", "job-1"), slog.String("mode", "code"))
	logger.InfoContext(ctx, "msg", "mode", "novel", "diff_bytes", 56676)

	out := append([]byte(nil), buf.Bytes()...)
	for _, key := range []string{"mode", "job_id"} {
		if n := countKey(t, out, key); n != 1 {
			t.Errorf("occurrences of %s = %d, want 1 (output: %s)", key, n, out)
		}
	}

	entries := decodeLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	// The more specific value wins: the context holds scope-wide values, the call site the value at hand.
	if entries[0]["mode"] != "novel" {
		t.Errorf("mode = %v, want novel (the call site's value)", entries[0]["mode"])
	}
	// Keys that do not collide still come from the context as before.
	if entries[0]["job_id"] != "job-1" {
		t.Errorf("job_id = %v, want job-1", entries[0]["job_id"])
	}
}

// When the same key is added twice, the one added later wins.
//
// Nesting With narrows the scope inward, so if a value meant to override on the inside lost
// to the outer one, there would be no way to override it.
func TestLaterWithWinsOnSameKey(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo)

	ctx := With(context.Background(), slog.String("phase", "collect"))
	logger.InfoContext(With(ctx, slog.String("phase", "publish")), "msg")

	out := append([]byte(nil), buf.Bytes()...)
	if n := countKey(t, out, "phase"); n != 1 {
		t.Fatalf("occurrences of phase = %d, want 1 (output: %s)", n, out)
	}
	entries := decodeLines(t, &buf)
	if entries[0]["phase"] != "publish" {
		t.Errorf("phase = %v, want publish (the value added later)", entries[0]["phase"])
	}
}

// Dropping collisions keeps the remaining attributes in order.
func TestSurvivingContextAttrsKeepOrder(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo)

	ctx := With(context.Background(),
		slog.String("job_id", "job-1"),
		slog.String("mode", "code"),
		slog.String("command", "review"),
	)
	logger.InfoContext(ctx, "msg", "mode", "novel")

	out := buf.Bytes()
	first, second := bytes.Index(out, []byte(`"job_id":`)), bytes.Index(out, []byte(`"command":`))
	if first < 0 || second < 0 || first > second {
		t.Errorf("job_id should come before command (output: %s)", out)
	}
}

// Collisions are collapsed inside a group too, since context attributes also go into the group.
func TestCallerAttrOverridesContextAttrInsideGroup(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(NewHandler(base).WithGroup("req"))

	ctx := With(context.Background(), slog.String("job_id", "job-1"))
	logger.InfoContext(ctx, "msg", "job_id", "job-2")

	out := append([]byte(nil), buf.Bytes()...)
	if n := countKey(t, out, "job_id"); n != 1 {
		t.Fatalf("occurrences of job_id = %d, want 1 (output: %s)", n, out)
	}
	entries := decodeLines(t, &buf)
	group, ok := entries[0]["req"].(map[string]any)
	if !ok {
		t.Fatalf("group req was not emitted: %v", entries[0])
	}
	if group["job_id"] != "job-2" {
		t.Errorf("req.job_id = %v, want job-2 (the call site's value)", group["job_id"])
	}
}

func TestWithNoAttrsReturnsSameContext(t *testing.T) {
	ctx := context.Background()
	if got := With(ctx); got != ctx {
		t.Error("With without attributes returned a different context")
	}
}

func TestAttrs(t *testing.T) {
	if got := Attrs(context.Background()); got != nil {
		t.Errorf("Attrs(empty context) = %v, want nil", got)
	}

	ctx := With(context.Background(), slog.String("job_id", "job-1"))
	attrs := Attrs(ctx)
	if len(attrs) != 1 || attrs[0].Key != "job_id" {
		t.Errorf("Attrs() = %v, want [job_id]", attrs)
	}

	//nolint:staticcheck // checks that passing a nil context does not panic
	if got := Attrs(nil); got != nil {
		t.Errorf("Attrs(nil) = %v, want nil", got)
	}
}

func TestHandlerRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	newTestLogger(&buf, slog.LevelWarn).Info("filtered out")
	if buf.Len() != 0 {
		t.Errorf("a log line below the level was emitted: %s", buf.String())
	}
}
