// Package jst は、日本標準時 (JST) への変換やフォーマットなど、
// JST を前提とした時刻処理を単純化するユーティリティ関数を提供します。
//
// このパッケージは表示層のための変換を担当します。永続化する時刻は UTC のまま扱い、
// 画面やログに出す直前で JST へ変換してください。
package jst

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// 表示に使うレイアウト。プロジェクト間で表記が割れないよう、ここに集約します。
const (
	// LayoutDisplay は履歴一覧などの分精度の表示に使います。
	// 末尾の MST はレイアウト指示子なので、実際のタイムゾーン略称 (JST) に置き換わります。
	LayoutDisplay = "2006-01-02 15:04 MST"

	// LayoutTimestamp は通知フッターなどの秒精度の表示に使います。
	// 末尾の JST はリテラルで、常に "JST" と出力されます。
	LayoutTimestamp = "2006/01/02 15:04:05 JST"
)

const locationName = "Asia/Tokyo"

// Location のキャッシュ。パッケージ変数の即時初期化にしないのは、ロード失敗時の警告を
// 利用側が slog のハンドラーを設定するより前に出力してしまうためです。
var (
	locationCache *time.Location
	locationOnce  sync.Once
)

// Location は、"Asia/Tokyo" の time.Location を一度だけロードし、そのポインタを返します。
// 二度目以降はキャッシュを返すため、呼び出しごとのファイルシステムへのアクセスは起きません。
func Location() *time.Location {
	locationOnce.Do(func() {
		locationCache = loadLocationOrFallback(locationName)
	})
	return locationCache
}

// loadLocationOrFallback は、指定された IANA タイムゾーン名の Location をロードします。
// ロードに失敗した場合は警告ログを出力し、UTC+9 の FixedZone にフォールバックします。
func loadLocationOrFallback(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		slog.Warn(
			"Failed to load location, falling back to FixedZone.",
			slog.String("location", name),
			slog.String("fallback", "FixedZone (UTC+9)"),
			slog.Any("error", err),
		)
		return time.FixedZone("JST", 9*60*60)
	}
	return loc
}

// Now は、日本標準時 (JST) における現在の時刻を返します。
func Now() time.Time {
	return time.Now().In(Location())
}

// From は、引数として渡された time.Time を JST に変換します。
func From(t time.Time) time.Time {
	return t.In(Location())
}

// Format は、与えられた time.Time を JST に変換した後、指定されたレイアウトでフォーマットします。
func Format(t time.Time, layout string) string {
	return From(t).Format(layout)
}

// FormatDisplay は LayoutDisplay で整形します。履歴一覧など、並べて表示する時刻の
// 表記を揃えるために使います。
//
// Format にレイアウト定数を渡しても同じ結果になりますが、2 つの定数は末尾の扱いが
// 異なるため取り違えても気づきにくく、名前で選べるようにしています。
func FormatDisplay(t time.Time) string {
	return Format(t, LayoutDisplay)
}

// FormatTimestamp は LayoutTimestamp で整形します。通知フッターなど、JST 固定である
// ことを明示したい箇所で使います。
func FormatTimestamp(t time.Time) string {
	return Format(t, LayoutTimestamp)
}

// Parse は、タイムゾーン情報を含まない時刻文字列を JST として解釈し、time.Time を返します。
//
// time.Parse はタイムゾーン略称をホストの time.Local と照合するため、実行環境の
// タイムゾーン設定次第でオフセットがずれます。Parse は Location を明示的に渡すことで
// この環境依存を排除します。
func Parse(value, layout string) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q as %q: %w", value, layout, err)
	}
	return t, nil
}
