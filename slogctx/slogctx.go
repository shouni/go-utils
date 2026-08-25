// Package slogctx は、context に積んだ属性を自動で付与する slog.Handler と、
// ログレベルの解決ヘルパーを提供します。
//
// リクエスト ID やジョブ ID のような「そのスコープのログすべてに載せたい値」を、
// 各ログ呼び出しへ引数で配って回らずに相関させるための仕組みです。
// 標準の slog.Logger.With はロガーを引き回す必要がありますが、こちらは context に
// 乗るため、既存の slog.XxxContext(ctx, ...) 呼び出しをそのまま相関ログにできます。
//
// 出力フォーマットには関与しません。GCP の Cloud Logging 向けに severity などを
// 詰め替える場合は、その HandlerOptions を持つハンドラーを NewHandler で包んでください。
package slogctx

import (
	"context"
	"log/slog"
	"slices"
	"strings"
)

// ParseLevel は環境変数などの文字列を slog のレベルへ変換します。
// 前後の空白は無視し、大文字小文字は区別しません。未知の値と空文字は Info とみなします。
func ParseLevel(raw string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type contextKey struct{}

// With は以降のログすべてに付与される属性を context に積みます。
// 積んだ属性は、NewHandler で包んだハンドラーがレコードへ自動的に追加します。
//
// 元の context が持つ属性は変更しないため、同じ context から分岐した処理同士が
// 互いの属性を汚染することはありません。
func With(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}

	existing := attrsFrom(ctx)
	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)

	return context.WithValue(ctx, contextKey{}, merged)
}

// Attrs は context に積まれた属性を返します。積まれていなければ nil を返します。
func Attrs(ctx context.Context) []slog.Attr {
	return attrsFrom(ctx)
}

func attrsFrom(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	attrs, _ := ctx.Value(contextKey{}).([]slog.Attr)
	return attrs
}

// NewHandler は、context に積まれた属性をレコードへ付与するハンドラーで base を包みます。
func NewHandler(base slog.Handler) slog.Handler {
	return &handler{Handler: base}
}

// handler は context 由来の属性をレコードへ付与する slog.Handler です。
type handler struct {
	slog.Handler
}

// Handle は context 由来の属性を足したうえで委譲先のハンドラーへ渡します。
//
// レコードが既に持っているキーは足しません。**呼び出し側が渡した値のほうが具体的**
// だからです（context に載るのはスコープ全体で共通の値で、呼び出し側はその場の値を
// 渡します）。
//
// ★ 落とさずに両方載せると、出力に同じキーが 2 つ並びます。JSON としては不正では
// ないため誰も失敗せず、**取り込む側の解釈に委ねられます。** Cloud Logging は連結
// するので、job_id が "job-1job-1" になります。**この形は検索に当たらないので、
// 失敗したジョブを ID で追えなくなります。**「相関のために載せた属性が、相関を
// 壊す」という裏返った結果になり、しかも壊れるのは呼び出し側も同じキーを載せた行
// ——つまり重要な行に偏ります。
//
// logger.With で足された属性はここからは見えない（委譲先が保持している）ため、
// そちらとの衝突は防げません。相関 ID は context に載せてください。
func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	attrs := attrsFrom(ctx)
	if len(attrs) == 0 {
		return h.Handler.Handle(ctx, record)
	}

	seen := make(map[string]struct{}, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		seen[a.Key] = struct{}{}
		return true
	})

	// context 自身の重複（同じキーで 2 度 With した場合）も潰します。**残すのは
	// 後から積んだほうです。** With を重ねるのはスコープを内側へ絞る操作なので、
	// 内側で上書きしたつもりの値が外側の値に負けると、上書きの手段が無くなります。
	// 後ろから見て初出だけを拾い、並びは元のまま戻します。
	kept := make([]slog.Attr, 0, len(attrs))
	for i := len(attrs) - 1; i >= 0; i-- {
		a := attrs[i]
		if _, dup := seen[a.Key]; dup {
			continue
		}
		seen[a.Key] = struct{}{}
		kept = append(kept, a)
	}
	slices.Reverse(kept)

	record.AddAttrs(kept...)
	return h.Handler.Handle(ctx, record)
}

// WithAttrs / WithGroup は委譲先を包み直し、context 属性の付与を維持します。
// 包み直さないと、logger.With(...) を通した時点で context 由来の属性が失われます。
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{Handler: h.Handler.WithGroup(name)}
}
