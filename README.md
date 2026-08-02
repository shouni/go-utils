# 📚 Go Utils

[![CI](https://github.com/shouni/go-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/shouni/go-utils/actions/workflows/ci.yml)
[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`Go Utils`** は、Go言語でアプリケーションを開発する際に繰り返し必要となる、汎用的なユーティリティ機能を集めたライブラリです。

このプロジェクトは、それぞれの機能が独立したパッケージとして提供されており、必要な機能のみをアプリケーションにインポートして利用することで、クリーンな依存関係を維持できます。

## ✨ 特徴

* **モジュール性**: 各ユーティリティが独立したGoパッケージとして提供されます。

### 収録基準 (What belongs here)

`utils` という名前は何でも受け入れてしまうため、収録の可否は以下で判断します。

1. **外部依存を持たない** — 標準ライブラリだけで完結すること。
   *現在の例外は `text` のみ（絵文字判定に `forPelevin/gomoji`、書記素クラスタ分割に `rivo/uniseg` を利用）。新規追加では認めません。*
2. **I/O やインフラに触れない** — ネットワーク・ファイルシステム・クラウドSDKを扱うものは対象外です。
   それらは `go-remote-io` や `gcp-kit` など、目的別のライブラリへ置いてください。
3. **2つ以上のプロジェクトから使われる** — 単一プロジェクトでしか使わないものは、
   その利用者側の `internal/` に置いてください。汎用に見えても、実際にはその
   プロジェクト固有の判断に紐づいていることが多いためです。

過去に基準を満たさないまま同居していた `giturl`（利用者が1つ・Git ドメイン固有）と
`iohandler`（利用者ゼロ）は v1.3.0 で削除しました。前者は
[git-gemini-web](https://github.com/shouni/git-gemini-web) の `internal/giturl` へ移設済みです。

---

## 🛠️ インストール

プロジェクト全体をインストールするには、以下のコマンドを使用します。

```bash
go get github.com/shouni/go-utils

```

特定のパッケージのみを利用する場合は、そのパッケージをインポートしてください（例: `import "github.com/shouni/go-utils/urlpath"`）。

---

## 📦 パッケージ構成 (Package Structure)

以下のパッケージがこのリポジトリで提供されています。

| パッケージ | 説明 | 主な提供機能 | 関連情報 |
| --- | --- | --- | --- |
| **`urlpath`** | **URLやリモートURI（GCS/S3）の解決**を行い、クラウドとローカルを透過的に扱います。 | クラウドURI判定 (`IsRemoteURI`)、パスの結合 (`ResolvePath`)、ディレクトリ解決 (`ResolveBaseDir`)、連番付与 (`GenerateIndexedPath`) | **リファクタ済** |
| **`envutil`** | **環境変数**の取得と型変換を安全に行うヘルパーを提供します。 | 環境変数取得 (`GetEnv`)、ブール値への変換 (`GetEnvAsBool`)、整数への変換 (`GetEnvAsInt`) | - |
| **`jst`** | **日本標準時 (JST) への変換**など、時刻処理を単純化します。表示層向けで、永続化する時刻は UTC のまま扱う想定です。 | JST現在時刻の取得 (`Now`)、任意の時刻をJSTへ変換 (`From`)、整形 (`Format`)、環境非依存のパース (`Parse`)、表示レイアウト定数 (`LayoutDisplay`, `LayoutTimestamp`) | v1.4.0 で `timeutil` から改名 |
| **`text`** | テキストデータのクリーンアップと整形を行います。 | 絵文字除去 (`CleanStringFromEmojis`)、**書記素クラスタ単位**の切詰め (`Truncate`)、リストパース | `forPelevin/gomoji` / `rivo/uniseg` 利用 |
| **`jobid`** | **非同期ジョブ識別子**の生成・検証・正規化を行います。ジョブ ID は URL パスとストレージパスの双方に現れるため、検証はセキュリティ境界を兼ねます。 | 検証 (`Validate`, `IsValid`)、パストラバーサル対策の正規化 (`Sanitize`)、時刻プレフィックス付き ID の生成 (`New`)、埋め込み時刻の復元 (`CreatedAt`) と並べ替えキー (`SortKey`) | 外部依存なし |
| **`slogctx`** | **context に積んだ属性を自動付与する `slog.Handler`** を提供します。リクエスト ID やジョブ ID を各ログ呼び出しへ配って回らずに相関できます。 | ログレベル解決 (`ParseLevel`)、属性の積み上げ (`With`, `Attrs`)、ハンドラーのラップ (`NewHandler`) | 外部依存なし・出力フォーマットには関与しない |

---

## 🚀 クイックスタート

### パスの解決 (`urlpath`)

```go
import "github.com/shouni/go-utils/urlpath"

// リモート(gs://等)かローカルかを問わず、適切にパスを結合します
path, _ := urlpath.ResolvePath("gs://my-bucket/images", "photo.png")
// path => "gs://my-bucket/images/photo.png"

```

### ジョブIDの検証と正規化 (`jobid`)

```go
import "github.com/shouni/go-utils/jobid"

// URLパスやストレージパスへ埋め込む前に、パス要素を落として検証します
id, err := jobid.Sanitize("../../20260726123456-abcd1234")
// id => "20260726123456-abcd1234"

// 埋め込まれた生成時刻を UTC で取り出します（表示は jst で JST へ変換）
createdAt, err := jobid.CreatedAt("video-recipe-20260726-123456-abcd1234")
// createdAt => 2026-07-26 12:34:56 +0000 UTC

// 用途プレフィックスが混在する一覧を、作成日時の降順で並べるためのキー
key := jobid.SortKey("video-recipe-20260726-123456-abcd1234")
// key => "20260726123456"（時刻を持たない ID では空文字）

```

### context を使ったログの相関 (`slogctx`)

```go
import "github.com/shouni/go-utils/slogctx"

// ハンドラーを包むと、context に積んだ属性が以降のログすべてに載ります
base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogctx.ParseLevel(os.Getenv("LOG_LEVEL"))})
slog.SetDefault(slog.New(slogctx.NewHandler(base)))

ctx := slogctx.With(ctx, slog.String("job_id", jobID))
slog.InfoContext(ctx, "phase started") // => {"job_id":"...", ...}

```

---

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

---
