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
func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	if attrs := attrsFrom(ctx); len(attrs) > 0 {
		record.AddAttrs(attrs...)
	}
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
