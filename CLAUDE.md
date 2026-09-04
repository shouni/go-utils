# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Go Utils (`github.com/shouni/go-utils`) is a Go library (not a service) holding only the small pieces that
were actually duplicated across several projects: `jobid` (async job identifiers), `jst` (Japan Standard
Time for display), `slogctx` (a `slog.Handler` that adds attributes carried on the context), and `strlist`
(normalising configured string lists). The four packages are independent — nothing here imports anything
else here.

**`go.mod` has no `require`, and there is no exception to that.** A name like `utils` accepts anything, so
admission is decided by three rules (also in README):

1. **No external dependency** — stdlib only.
2. **No I/O or infrastructure** — network, filesystem and cloud SDKs belong in a purpose-built library
   (`go-remote-io`, `gcp-kit`, …).
3. **Used by two or more projects** — anything used by one lives in that project's `internal/`. Code that
   looks general is usually tied to one project's judgement.

## Commands

```bash
go build ./...
go vet ./...
gofmt -l .                 # must print nothing (CI fails otherwise)
go test -race ./...        # what CI runs
go test ./jobid -run TestValidate -v      # single test
golangci-lint run ./...
go run golang.org/x/exp/cmd/gorelease@latest   # run before tagging
```

CI (`.github/workflows/ci.yml`) is a thin caller of the shared
`shouni/workflows/.github/workflows/go-ci.yml@v1`; no fuzz targets are declared.

## Package notes and invariants

### `jobid`

- **Validation doubles as a security boundary.** A job ID appears both in an HTTP route parameter and in an
  object-storage path, and several services exchange the same IDs — if "what counts as a valid ID" differs
  between them, one rejects what the other issued, or only one of them allows traversal. That decision is
  centralised here.
- **The grammar is: starts alphanumeric, then alphanumerics / `-` / `_`, up to `MaxLength` (128).**
  Restricting the first character keeps values that start with `-` or `_` from being reinterpreted as
  command-line flags or query syntax; the character set structurally excludes `/`, `..` and percent-encoding.
  All allowed characters are ASCII, so byte-wise scanning is correct and length equals character count.
- **`Sanitize` rejects empty input before `path.Base`**, because `path.Base("")` returns `"."` and a `.`
  that was never in the input misdirects whoever reads the error.
- **Failures are typed** (`ErrEmpty` / `ErrTooLong` / `ErrInvalidFormat`, and `ErrNoTimestamp` for
  `CreatedAt`) and classified with `errors.Is`. They all become 400 in a handler, but the kinds stay
  separate so response text and metric labels can differ. Do not make callers compare message strings.
- **`New` never fails over a bad prefix.** The prefix is decoration; uniqueness comes from the timestamp and
  the random suffix, so `normalizePrefix` strips disallowed characters and truncates rather than erroring —
  failing to issue an ID because of a prefix's spelling is the worse outcome. Timestamps are UTC.
- **Lexicographic order equals newest-first only when every ID in the list shares one prefix.** That is the
  condition under which an object-storage listing needs no separate index; mixed-prefix listings must sort
  by `SortKey`.
- **`CreatedAt` still reads two pre-`New` formats** (`c20060102-150405-…` and `20060102150405-…`) because
  old IDs live on in artifact paths. It also rejects timestamps before `minTimestamp` (2000-01-01): random
  hex can produce a run of 14 digits that passes plain range checks.

### `slogctx`

- **Correlation IDs belong on the context, not on a logger.** `slog.Logger.With` requires threading the
  logger through; `slogctx.With` rides the context, so existing `slog.XxxContext(ctx, …)` calls become
  correlated without touching them.
- **Key collisions are resolved deliberately, and both rules matter.** A key already on the record wins
  over the context (the call site's value is more specific); within the context, the later `With` wins
  (nesting narrows scope). Attributes added via `logger.With` are held by the delegate and are invisible
  here, so collisions with those cannot be prevented — put correlation IDs on the context.
- **Why deduplicate at all:** emitting both copies is valid JSON, so nothing fails; Cloud Logging
  concatenates them and `job_id` becomes `"job-1job-1"`, which no search matches. The attribute added for
  correlation breaks correlation, and it breaks precisely on the lines where the call site cared enough to
  pass the key itself.
- **`WithAttrs` / `WithGroup` re-wrap the delegate.** Without that, one `logger.With(...)` drops context
  attributes from then on.
- **The package does not touch output format.** For Cloud Logging's `severity`/`message` keys, wrap a
  handler that carries those `HandlerOptions` — that is `gcp-kit/cloudlog`'s job.
- `ParseLevel` delegates to `slog.Level.UnmarshalText` (so `"DEBUG+2"` works), maps `"WARNING"` to `WARN`
  because it is common in env vars, and treats unknown/empty as Info.

### `jst`

- **Display-layer only.** Persist times as UTC and convert at the moment of rendering.
- **The `Asia/Tokyo` location is loaded lazily via `sync.OnceValue`, not in a package-level initialiser** —
  eager loading would emit its failure warning before the consumer has installed its slog handler. A load
  failure falls back to a `FixedZone("JST", +9h)`.
- Layout constants live here so the notation cannot drift between projects: `LayoutDisplay` ends in `MST`
  (a layout directive, replaced by the real abbreviation), `LayoutTimestamp` ends in a literal `JST`.

### `strlist`

- **It normalises an already-split list; splitting is the config library's job** (`caarlos0/env` and the
  like). `strings.Split` leaves `"a, b,,a"` as `["a", " b", "", "a"]`, which makes allowlist comparisons
  miss on a padded element and menus show a duplicate.
- **Empty in, empty out — never a default.** Whether empty is acceptable is the caller's validation
  decision; substituting a default here would hide a missing setting.
- `NormalizeFold` lowercases and treats case as duplication — for hostnames, mail domains and similar
  case-insensitive settings (the comparison side must lowercase too). Use `Normalize` where case carries
  meaning (tokens, API keys).

## Testing notes

- **No assertion library, and `go.mod` must stay empty.** A test-only requirement still lands in every
  consumer's `go.sum`, and the empty `go.sum` is itself the thing being protected.
- `jobid` and `slogctx` test in-package (they reach `newAt` and the handler internals); `jst` and `strlist`
  test as external `_test` packages, with `jst` keeping an extra `internal_test.go` for
  `loadLocationOrFallback`.
- `jobid`, `jst` and `strlist` carry `example_test.go` with `// Output:` comments, so the examples run as
  tests — keep them deterministic.
- Doc comments and comments in this repo are Japanese; error strings are English (lowercase, no trailing
  punctuation). Consumers classify with `errors.Is`, and tests must not assert on message substrings.
