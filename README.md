# 📚 Go Utils

[![CI](https://github.com/shouni/go-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/shouni/go-utils/actions/workflows/ci.yml)
[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)
[![Language](https://img.shields.io/badge/Language-Go-blue)](https://go.dev/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://go.dev/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/shouni/go-utils.svg)](https://pkg.go.dev/github.com/shouni/go-utils)

## 🚀 概要 (About) - 複数プロジェクトで重複していた、標準ライブラリだけの小物

**`Go Utils`** は、複数のプロジェクトで実際に重複していた小さな処理だけを集めたモジュールです。パッケージ同士は独立しているため、必要なものだけをインポートできます。

## ✨ 収録基準 (What belongs here)

`utils` という名前は何でも受け入れてしまうため、収録の可否は以下で判断します。

1. **外部依存を持たない** — 標準ライブラリだけで完結すること。
   *現在このモジュールの `go.mod` に `require` はありません。例外は作らない方針です。*
2. **I/O やインフラに触れない** — ネットワーク・ファイルシステム・クラウドSDKを扱うものは対象外です。
   それらは `go-remote-io` や `gcp-kit` など、目的別のライブラリへ置いてください。
3. **2つ以上のプロジェクトから使われる** — 単一プロジェクトでしか使わないものは、
   その利用者側の `internal/` に置いてください。汎用に見えても、実際にはそのプロジェクト固有の判断に紐づいていることが多いためです。

## 🛠️ インストール (Installation)

```bash
go get github.com/shouni/go-utils
```

インポートはパッケージ単位で行います（例: `import "github.com/shouni/go-utils/jobid"`）。

## 📦 パッケージ構成 (Package Structure)

| パッケージ | 説明 | 主な提供機能 |
| --- | --- | --- |
| **`jobid`** | **非同期ジョブ識別子**の生成・検証・正規化を行います。ジョブ ID は URL パスとストレージパスの双方に現れるため、検証はセキュリティ境界を兼ねます。 | 検証 (`Validate`, `IsValid`) と種類別のエラー (`ErrEmpty`, `ErrTooLong`, `ErrInvalidFormat`)、パストラバーサル対策の正規化 (`Sanitize`)、用途プレフィックスと生成時刻を含む ID の採番 (`New`)、埋め込み時刻の復元 (`CreatedAt`) と並べ替えキー (`SortKey`) |
| **`slogctx`** | **context に積んだ属性を自動付与する `slog.Handler`** を提供します。リクエスト ID やジョブ ID を各ログ呼び出しへ配って回らずに相関できます。出力フォーマットには関与しません。 | ログレベル解決 (`ParseLevel`)、属性の積み上げ (`With`, `Attrs`)、ハンドラーのラップ (`NewHandler`) |
| **`jst`** | **日本標準時 (JST) への変換**など、時刻処理を単純化します。表示層向けで、永続化する時刻は UTC のまま扱う想定です。 | 現在時刻の取得 (`Now`)、任意の時刻を JST へ変換 (`From`)、整形 (`Format`) と定数つきの近道 (`FormatDisplay`, `FormatTimestamp`)、環境非依存のパース (`Parse`)、ロケーション取得 (`Location`)、表示レイアウト定数 (`LayoutDisplay`, `LayoutTimestamp`) |
| **`strlist`** | 設定値として読み込んだ**分割済みの文字列リスト**を整えます。カンマ区切りの分割そのものは設定ライブラリの担当で、その後始末を引き受けます。 | 前後の空白・空要素・重複を落とす正規化 (`Normalize`)、大文字小文字を区別しない正規化 (`NormalizeFold`) |

## 🚀 クイックスタート (Quick Start)

### ジョブ ID の検証と正規化 (`jobid`)

```go
import "github.com/shouni/go-utils/jobid"

// URLパスやストレージパスへ埋め込む前に、パス要素を落として検証します
id, err := jobid.Sanitize("../../20260725123456-abcd1234")
// id => "20260725123456-abcd1234"

// 埋め込まれた生成時刻を UTC で取り出します（表示は jst で JST へ変換）
createdAt, err := jobid.CreatedAt("video-recipe-20260725-150405-a1b2c3d4")
// createdAt => 2026-07-25 15:04:05 +0000 UTC

// 用途プレフィックスが混在する一覧を、作成日時の降順で並べるためのキー
key := jobid.SortKey("video-recipe-20260725-150405-a1b2c3d4")
// key => "20260725150405"（時刻を持たない ID では空文字）
```

検証の失敗は種類ごとに `errors.Is` で判定できます。

```go
switch {
case errors.Is(err, jobid.ErrEmpty):         // 空（または空白のみ）
case errors.Is(err, jobid.ErrTooLong):       // MaxLength 超過
case errors.Is(err, jobid.ErrInvalidFormat): // 使えない文字を含む
}
```

### context を使ったログの相関 (`slogctx`)

```go
import (
	"log/slog"
	"os"

	"github.com/shouni/go-utils/slogctx"
)

// ハンドラーを包むと、context に積んだ属性が以降のログすべてに載ります
base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogctx.ParseLevel(os.Getenv("LOG_LEVEL"))})
slog.SetDefault(slog.New(slogctx.NewHandler(base)))

ctx = slogctx.With(ctx, slog.String("job_id", jobID))
slog.InfoContext(ctx, "phase started") // => {"job_id":"...", ...}

// 呼び出し側が同じキーを渡した場合は、その場の値のほうが具体的なので呼び出し側だけが載ります
slog.InfoContext(ctx, "phase started", "job_id", "other") // => {"job_id":"other", ...}
```

同じキーを 2 度積んだ場合は後から積んだほうが残ります。ただし `logger.With` で足した属性は
委譲先が保持していてここからは見えないため、そちらとの衝突は防げません。相関 ID は
context に載せてください。衝突時の扱いとその理由は `NewHandler` のコメントにあります。

### 時刻の表示 (`jst`)

```go
import "github.com/shouni/go-utils/jst"

// 永続化された UTC の時刻を、表示の直前で JST へ変換します
jst.FormatDisplay(createdAt)   // => "2026-07-25 15:04 JST"   一覧向け（分精度）
jst.FormatTimestamp(createdAt) // => "2026/07/25 15:04:05 JST" 通知フッター向け（秒精度）
```

`jobid` / `jst` / `strlist` には `example_test.go`（`go test` で出力まで検証される実行可能な例）があります。詳しい使い方はそちらを参照してください。

## 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。
