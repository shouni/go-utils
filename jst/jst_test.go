package jst_test

import (
	"errors"
	"testing"
	"time"

	"github.com/shouni/go-utils/jst"
)

// jstOffset は UTC と JST のオフセット (9時間) です。
const jstOffset = 9 * time.Hour

// isJST は、ロケーション名が JST 相当 ("Asia/Tokyo" またはフォールバックの "JST") かを判定します。
func isJST(name string) bool {
	return name == "Asia/Tokyo" || name == "JST"
}

// TestLocation は、Location() が正しいロケーションを返し、かつキャッシュされることを確認します。
func TestLocation(t *testing.T) {
	loc1 := jst.Location()
	if !isJST(loc1.String()) {
		t.Errorf("Location().String() = %q, want \"Asia/Tokyo\" or \"JST\"", loc1.String())
	}

	// sync.Once によるキャッシュが効いていれば、ポインタが同一になる。
	if loc2 := jst.Location(); loc1 != loc2 {
		t.Error("Location() が呼び出しごとに異なるポインタを返しました。キャッシュされていません")
	}

	if _, offset := time.Now().In(loc1).Zone(); offset != int(jstOffset.Seconds()) {
		t.Errorf("Location() のオフセット = %d, want %d", offset, int(jstOffset.Seconds()))
	}
}

// TestNow は、現在時刻が JST (UTC+9) のタイムゾーン情報を持って取得されることを確認します。
func TestNow(t *testing.T) {
	now := jst.Now()

	if !isJST(now.Location().String()) {
		t.Errorf("Now().Location() = %q, want \"Asia/Tokyo\" or \"JST\"", now.Location().String())
	}

	if _, offset := now.Zone(); offset != int(jstOffset.Seconds()) {
		t.Errorf("Now() のオフセット = %d, want %d", offset, int(jstOffset.Seconds()))
	}

	// 絶対時刻が現在時刻に近いこと (実行速度を考慮し1秒の誤差を許容)。
	if diff := time.Since(now); diff > time.Second || diff < -time.Second {
		t.Errorf("Now() が現在の絶対時刻とずれています。Diff: %v", diff)
	}
}

// TestFrom は、UTC 時刻を JST へ正しく変換することを確認します。
func TestFrom(t *testing.T) {
	utc := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	want := time.Date(2025, time.January, 1, 9, 0, 0, 0, jst.Location())

	got := jst.From(utc)

	// 絶対時刻が変わっていないこと。
	if !got.Equal(want) {
		t.Errorf("From(%v) = %v, want %v", utc, got, want)
	}
	if got.Location() != jst.Location() {
		t.Errorf("From(%v).Location() = %q, want %q", utc, got.Location(), jst.Location())
	}
}

// TestFormat は、JST へ変換した上で指定レイアウトに整形されることを確認します。
func TestFormat(t *testing.T) {
	utc := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		layout string
		want   string
	}{
		{"任意のレイアウト", "2006/01/02 15:04", "2025/01/01 09:00"},
		{"LayoutDisplay", jst.LayoutDisplay, "2025-01-01 09:00 JST"},
		{"LayoutTimestamp", jst.LayoutTimestamp, "2025/01/01 09:00:00 JST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jst.Format(utc, tt.layout); got != tt.want {
				t.Errorf("Format(%v, %q) = %q, want %q", utc, tt.layout, got, tt.want)
			}
		})
	}
}

// TestParse は、タイムゾーン情報を含まない文字列が JST として解釈されることを確認します。
func TestParse(t *testing.T) {
	got, err := jst.Parse("2025-01-01 18:30", "2006-01-02 15:04")
	if err != nil {
		t.Fatalf("Parse() が予期しないエラーを返しました: %v", err)
	}

	want := time.Date(2025, time.January, 1, 18, 30, 0, 0, jst.Location())
	if !got.Equal(want) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}

	// ホストの time.Local に依存せず、必ず UTC+9 として解釈されること。
	if _, offset := got.Zone(); offset != int(jstOffset.Seconds()) {
		t.Errorf("Parse() のオフセット = %d, want %d", offset, int(jstOffset.Seconds()))
	}
}

// TestParse_Invalid は、不正な入力に対してエラーを返し、元のエラーをラップすることを確認します。
func TestParse_Invalid(t *testing.T) {
	_, err := jst.Parse("25:00", "15:04")
	if err == nil {
		t.Fatal("Parse() が不正な入力に対してエラーを返しませんでした")
	}

	// %w でラップされ、time.ParseError を取り出せること。
	if _, ok := errors.AsType[*time.ParseError](err); !ok {
		t.Errorf("Parse() のエラーが *time.ParseError をラップしていません: %v", err)
	}
}

// TestFormatDisplayAndTimestamp は、レイアウト定数を直接渡した場合と同じ結果になり、
// 2 つの定数の違い（分精度とタイムゾーン略称 / 秒精度と "JST" リテラル）が出ることを確認します。
func TestFormatDisplayAndTimestamp(t *testing.T) {
	utcTime := time.Date(2026, time.July, 25, 6, 4, 5, 0, time.UTC)

	if got, want := jst.FormatDisplay(utcTime), jst.Format(utcTime, jst.LayoutDisplay); got != want {
		t.Errorf("FormatDisplay() = %q, want %q", got, want)
	}
	if got, want := jst.FormatTimestamp(utcTime), jst.Format(utcTime, jst.LayoutTimestamp); got != want {
		t.Errorf("FormatTimestamp() = %q, want %q", got, want)
	}

	if got, want := jst.FormatDisplay(utcTime), "2026-07-25 15:04 JST"; got != want {
		t.Errorf("FormatDisplay() = %q, want %q", got, want)
	}
	if got, want := jst.FormatTimestamp(utcTime), "2026/07/25 15:04:05 JST"; got != want {
		t.Errorf("FormatTimestamp() = %q, want %q", got, want)
	}
}
