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
//
// 解釈は slog.Level.UnmarshalText に委ねるため、"DEBUG" などの名前に加えて
// slog が定める相対表記（"DEBUG+2" や "ERROR-1"）もそのまま使えます。
// "WARNING" だけは slog が受け付けないものの、環境変数としては広く使われるため
// "WARN" と同義に扱います。
func ParseLevel(raw string) slog.Level {
	name := strings.ToUpper(strings.TrimSpace(raw))
	if name == "WARNING" {
		name = "WARN"
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(name)); err != nil {
		return slog.LevelInfo
	}
	return level
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
//
// キーが衝突したときの扱いは 2 つです。レコードが既に持っているキーは context から足さず
// （呼び出し側が渡したその場の値のほうが具体的なため）、context 自身の重複は後から積んだ
// ほうを残します（With を重ねるのはスコープを内側へ絞る操作のため）。
// ただし logger.With で足された属性は委譲先が保持していてここからは見えないため、
// そちらとの衝突は防げません。相関 ID は context に載せてください。
func NewHandler(base slog.Handler) slog.Handler {
	return &handler{Handler: base}
}

type handler struct {
	slog.Handler
}

// Handle は context 由来の属性を足したうえで委譲先のハンドラーへ渡します。
//
// 衝突したキーを落とさずに両方載せると、出力に同じキーが 2 つ並びます。JSON としては
// 不正ではないため誰も失敗せず、解釈は取り込む側任せになります。Cloud Logging は連結する
// ので job_id が "job-1job-1" となり、その ID で検索しても当たりません。相関のために載せた
// 属性が相関を壊すうえ、壊れるのは呼び出し側も同じキーを載せた行、つまり重要な行に偏ります。
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

	// context 自身の重複は後から積んだほうを残すため、後ろから見て初出だけを拾い、
	// 並びは元のまま戻します。
	kept := make([]slog.Attr, 0, len(attrs))
	for _, a := range slices.Backward(attrs) {
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
