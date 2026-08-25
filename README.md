# 📚 Go Utils

[![CI](https://github.com/shouni/go-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/shouni/go-utils/actions/workflows/ci.yml)
[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)
[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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
| **`jobid`** | **非同期ジョブ識別子**の生成・検証・正規化を行います。ジョブ ID は URL パスとストレージパスの双方に現れるため、検証はセキュリティ境界を兼ねます。 | 検証 (`Validate`, `IsValid`)、パストラバーサル対策の正規化 (`Sanitize`)、用途プレフィックスと生成時刻を含む ID の採番 (`New`)、埋め込み時刻の復元 (`CreatedAt`) と並べ替えキー (`SortKey`) |
| **`slogctx`** | **context に積んだ属性を自動付与する `slog.Handler`** を提供します。リクエスト ID やジョブ ID を各ログ呼び出しへ配って回らずに相関できます。出力フォーマットには関与しません。 | ログレベル解決 (`ParseLevel`)、属性の積み上げ (`With`, `Attrs`)、ハンドラーのラップ (`NewHandler`) |
| **`jst`** | **日本標準時 (JST) への変換**など、時刻処理を単純化します。表示層向けで、永続化する時刻は UTC のまま扱う想定です。 | 現在時刻の取得 (`Now`)、任意の時刻を JST へ変換 (`From`)、整形 (`Format`)、環境非依存のパース (`Parse`)、ロケーション取得 (`Location`)、表示レイアウト定数 (`LayoutDisplay`, `LayoutTimestamp`) |
| **`strlist`** | 設定値として読み込んだ**分割済みの文字列リスト**を整えます。カンマ区切りの分割そのものは設定ライブラリの担当で、その後始末を引き受けます。 | 前後の空白・空要素・重複を落とす正規化 (`Normalize`) |

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

// 呼び出し側が同じキーを渡した場合は、呼び出し側の値だけが載ります（その場の値の
// ほうが具体的なため）。同じキーを 2 つ並べると、JSON としては不正ではないので
// 誰も失敗せず、解釈は取り込む側任せになります。Cloud Logging は連結するため
// job_id が "job-1job-1" となり、その ID で検索しても当たらなくなります。
slog.InfoContext(ctx, "phase started", "job_id", "other") // => {"job_id":"other", ...}
```

同じキーを 2 度積んだ場合は、**後から積んだほうが残ります**（`With` を重ねるのは
スコープを内側へ絞る操作なので、内側の上書きが効かないと手段が無くなります）。
`logger.With` で足した属性は委譲先が保持していてここからは見えないため、そちらとの
衝突は防げません。相関 ID は context に載せてください。

`jobid` / `jst` / `strlist` には `example_test.go`（`go test` で出力まで検証される実行可能な例）があります。詳しい使い方はそちらを参照してください。

## 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。
