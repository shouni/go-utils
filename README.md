# 📚 Go Utils

[![CI](https://github.com/shouni/go-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/shouni/go-utils/actions/workflows/ci.yml)
[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)
[![Language](https://img.shields.io/badge/Language-Go-blue)](https://go.dev/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://go.dev/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/shouni/go-utils.svg)](https://pkg.go.dev/github.com/shouni/go-utils)

English | [日本語](README.ja.md)

## 🚀 About - small, stdlib-only pieces that were duplicated across several projects

**`Go Utils`** collects only the small pieces of code that were actually duplicated across several projects. The packages are independent of each other, so you can import just the ones you need.

## ✨ What belongs here

A name like `utils` accepts anything, so admission is decided by the following rules.

1. **No external dependencies** — everything must work with the standard library alone.
   *This module's `go.mod` currently has no `require`, and the policy is to make no exceptions.*
2. **No I/O or infrastructure** — anything dealing with the network, the filesystem or cloud SDKs is out of scope.
   Put it in a purpose-built library such as `go-remote-io` or `gcp-kit`.
3. **Used by two or more projects** — anything used by only one project belongs in
   that project's `internal/`. Code that looks general is usually tied to one project's own judgement.

## 📦 Package Structure

| Package | Description | Main features |
| --- | --- | --- |
| **`jobid`** | Generates, validates and normalises **asynchronous job identifiers**. A job ID appears both in URL paths and in storage paths, so validation doubles as a security boundary. | Validation (`Validate`, `IsValid`) with per-kind errors (`ErrEmpty`, `ErrTooLong`, `ErrInvalidFormat`), normalisation against path traversal (`Sanitize`), minting IDs with a purpose prefix and creation time (`New`), recovering the embedded time (`CreatedAt`) and a sort key (`SortKey`) |
| **`slogctx`** | Provides **a `slog.Handler` that automatically adds attributes carried on the context**. Request IDs and job IDs can be correlated without handing them to every logging call. It does not touch the output format. | Log level resolution (`ParseLevel`), adding attributes (`With`, `Attrs`), wrapping a handler (`NewHandler`) |
| **`jst`** | Simplifies time handling such as **converting to Japan Standard Time (JST)**. It is meant for the display layer; persisted times are expected to stay in UTC. | Current time (`Now`), converting any time to JST (`From`), formatting (`Format`) with shortcuts for the constants (`FormatDisplay`, `FormatTimestamp`), environment-independent parsing (`Parse`), the location (`Location`), display layout constants (`LayoutDisplay`, `LayoutTimestamp`) |
| **`paging`** | **Page metadata for listing responses** (`PageMeta`) and building it from a total count (`New`). The common type that lets GCS-backed and Firestore-backed listings return the same JSON shape. | Building (`New`), the number of items to skip (`Offset`), adjusting to the number of items actually fetched (`WithItemCount`) |
| **`strlist`** | Tidies **already-split string lists** read in as configuration values. Splitting on commas is the configuration library's job; this handles the cleanup after it. | Normalisation that drops surrounding whitespace, empty elements and duplicates (`Normalize`), case-insensitive normalisation (`NormalizeFold`) |

## 🚦 Usage

Install with `go get github.com/shouni/go-utils` and import per package
(e.g. `import "github.com/shouni/go-utils/jobid"`).

### Validating and normalising job IDs (`jobid`)

```go
import "github.com/shouni/go-utils/jobid"

// Drop path elements and validate before embedding the ID in a URL or storage path
id, err := jobid.Sanitize("../../20260725123456-abcd1234")
// id => "20260725123456-abcd1234"

// Extract the embedded creation time in UTC (convert to JST with jst for display)
createdAt, err := jobid.CreatedAt("video-recipe-20260725-150405-a1b2c3d4")
// createdAt => 2026-07-25 15:04:05 +0000 UTC

// A key for sorting a list with mixed purpose prefixes newest-first by creation time
key := jobid.SortKey("video-recipe-20260725-150405-a1b2c3d4")
// key => "20260725150405" (empty for an ID without a timestamp)
```

Validation failures can be classified by kind with `errors.Is`.

```go
switch {
case errors.Is(err, jobid.ErrEmpty):         // empty (or whitespace only)
case errors.Is(err, jobid.ErrTooLong):       // exceeds MaxLength
case errors.Is(err, jobid.ErrInvalidFormat): // contains a disallowed character
}
```

### Correlating logs through the context (`slogctx`)

```go
import (
	"log/slog"
	"os"

	"github.com/shouni/go-utils/slogctx"
)

// Once the handler is wrapped, attributes carried on the context appear on every later log line
base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogctx.ParseLevel(os.Getenv("LOG_LEVEL"))})
slog.SetDefault(slog.New(slogctx.NewHandler(base)))

ctx = slogctx.With(ctx, slog.String("job_id", jobID))
slog.InfoContext(ctx, "phase started") // => {"job_id":"...", ...}

// When the call site passes the same key, its value is more specific, so only the call site's value appears
slog.InfoContext(ctx, "phase started", "job_id", "other") // => {"job_id":"other", ...}
```

When the same key is added twice, the one added later wins. Attributes added with `logger.With`,
however, are held by the delegate handler and invisible here, so collisions with them cannot be
prevented. Put correlation IDs on the context. How collisions are handled, and why, is explained in
the comment on `NewHandler`.

### Displaying times (`jst`)

```go
import "github.com/shouni/go-utils/jst"

// Convert a persisted UTC time to JST just before display
jst.FormatDisplay(createdAt)   // => "2026-07-25 15:04 JST"   for listings (minute precision)
jst.FormatTimestamp(createdAt) // => "2026/07/25 15:04:05 JST" for notification footers (second precision)
```

`jobid` / `jst` / `strlist` have an `example_test.go` (runnable examples whose output is verified by `go test`). See them for more detailed usage.

## 📜 License

This project is released under the [MIT License](https://opensource.org/licenses/MIT).
