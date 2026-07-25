package slogctx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

// newTestLogger は、JSON 出力を buf へ書く context 対応ロガーを返します。
func newTestLogger(buf *bytes.Buffer, level slog.Level) *slog.Logger {
	return slog.New(NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})))
}

// decodeLines は出力された JSON ログを 1 行ずつマップへ復号します。
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
		// 環境変数は前後に空白が混ざりやすいため、トリムされること。
		{" warn ", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"", slog.LevelInfo},
		{"nonsense", slog.LevelInfo},
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
		t.Errorf("entry[0] = %v, want job_id/command を含む", entries[0])
	}
	if _, ok := entries[1]["job_id"]; ok {
		t.Errorf("属性を積んでいない context のログに job_id が漏れている: %v", entries[1])
	}
}

// 同じ context から分岐した処理同士が互いの属性を汚染しないこと。
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

// logger.With で包み直したあとも context 由来の属性が消えないこと。
func TestContextAttrsSurviveLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, slog.LevelInfo).With("component", "pipeline")

	logger.InfoContext(With(context.Background(), slog.String("job_id", "job-1")), "msg")

	entries := decodeLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0]["component"] != "pipeline" || entries[0]["job_id"] != "job-1" {
		t.Errorf("entry = %v, want component/job_id を両方含む", entries[0])
	}
}

// WithGroup で包み直したあとも context 由来の属性が消えないこと。
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
		t.Fatalf("グループ req が出力されていない: %v", entries[0])
	}
	if group["job_id"] != "job-1" {
		t.Errorf("req.job_id = %v, want job-1", group["job_id"])
	}
}

func TestWithNoAttrsReturnsSameContext(t *testing.T) {
	ctx := context.Background()
	if got := With(ctx); got != ctx {
		t.Error("属性なしの With が別の context を返している")
	}
}

func TestAttrs(t *testing.T) {
	if got := Attrs(context.Background()); got != nil {
		t.Errorf("Attrs(空 context) = %v, want nil", got)
	}

	ctx := With(context.Background(), slog.String("job_id", "job-1"))
	attrs := Attrs(ctx)
	if len(attrs) != 1 || attrs[0].Key != "job_id" {
		t.Errorf("Attrs() = %v, want [job_id]", attrs)
	}

	//nolint:staticcheck // nil context を渡しても panic しないことの確認。
	if got := Attrs(nil); got != nil {
		t.Errorf("Attrs(nil) = %v, want nil", got)
	}
}

func TestHandlerRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	newTestLogger(&buf, slog.LevelWarn).Info("filtered out")
	if buf.Len() != 0 {
		t.Errorf("レベル未満のログが出力された: %s", buf.String())
	}
}
