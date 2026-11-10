// Package slogctx provides a slog.Handler that automatically adds attributes carried on
// the context, and a helper for resolving log levels.
//
// It correlates values that belong on every log line in a scope, such as a request ID or
// job ID, without passing them as arguments to each logging call. The standard
// slog.Logger.With requires threading the logger through; these attributes ride the
// context instead, so existing slog.XxxContext(ctx, ...) calls become correlated as they are.
//
// It does not touch the output format. To remap keys such as severity for GCP Cloud Logging,
// wrap a handler carrying those HandlerOptions with NewHandler.
package slogctx

import (
	"context"
	"log/slog"
	"slices"
	"strings"
)

// ParseLevel converts a string, such as an environment variable, to a slog level.
// Surrounding whitespace is ignored and case does not matter. Unknown values and the
// empty string are treated as Info.
//
// Parsing is delegated to slog.Level.UnmarshalText, so besides names such as "DEBUG",
// slog's relative notation ("DEBUG+2", "ERROR-1") works as is.
// "WARNING" is not accepted by slog, but it is common in environment variables, so it is
// treated as a synonym for "WARN".
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

// With adds attributes to the context that are attached to every subsequent log record.
// A handler wrapped with NewHandler adds them to each record automatically.
//
// The attributes held by the parent context are not modified, so work branched from the
// same context never pollutes each other's attributes.
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

// Attrs returns the attributes carried on the context, or nil if there are none.
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

// NewHandler wraps base in a handler that adds the context's attributes to each record.
//
// Key collisions are resolved by two rules. A key the record already has is not added
// from the context (the value the call site passed is more specific), and among the
// context's own duplicates the one added later wins (nesting With narrows the scope).
// Attributes added with logger.With, however, are held by the delegate and invisible here,
// so collisions with them cannot be prevented. Put correlation IDs on the context.
func NewHandler(base slog.Handler) slog.Handler {
	return &handler{Handler: base}
}

type handler struct {
	slog.Handler
}

// Handle adds the context's attributes and passes the record to the delegate handler.
//
// Emitting both copies of a colliding key puts the same key twice in the output. That is
// still valid JSON, so nothing fails and interpretation is left to the ingesting side.
// Cloud Logging concatenates them, so job_id becomes "job-1job-1" and a search for that ID
// finds nothing. The attribute added for correlation breaks correlation, and it breaks on
// the lines where the call site passed the same key too, which are the important ones.
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

	// Among the context's own duplicates the later one wins, so walk from the back keeping
	// only the first occurrence, then restore the original order.
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

// WithAttrs / WithGroup re-wrap the delegate so context attributes keep being added.
// Without re-wrapping, the context's attributes would be lost as soon as logger.With(...) is used.
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{Handler: h.Handler.WithGroup(name)}
}
