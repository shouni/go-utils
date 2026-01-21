# 📚 Go Utils

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`Go Utils`** は、Go言語でアプリケーションを開発する際に繰り返し必要となる、汎用的で再利用可能なユーティリティ機能を集めたライブラリです。

このプロジェクトは、それぞれの機能が独立したパッケージとして提供されており、必要な機能のみをアプリケーションにインポートして利用することで、クリーンな依存関係を維持できます。

## ✨ 特徴

* **モジュール性**: 各ユーティリティが独立したGoパッケージとして提供されます。
* **安全性**: SSRF対策やセキュアなURL検証など、セキュリティを考慮した設計です。
* **クラウド対応**: GCSやS3などのリモートURIとローカルパスを透過的に扱うためのツールが含まれています。
* **高信頼性**: 外部ライブラリ（例: `cenkalti/backoff`）をラップし、より使いやすいインターフェースを提供します。

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
| **`security`** | **セキュリティ検証**を提供します。SSRF対策や、サービスURLの安全性を判定します。 | SSRF対策のURL検証 (`IsSafeURL`)、Secure属性向けのURL判定 (`IsSecureServiceURL`) | **新規追加** |
| **`urlpath`** | **URLやリモートURI（GCS/S3）の解決**を行い、クラウドとローカルを透過的に扱います。 | クラウドURI判定 (`IsRemoteURI`)、パスの結合 (`ResolvePath`)、ディレクトリ解決 (`ResolveBaseDir`)、連番付与 (`GenerateIndexedFiles`) | **リファクタ済** |
| **`iohandler`** | **ファイルI/Oと標準入出力を抽象化**します。CLIアプリケーションでの処理を簡潔にします。 | ファイル/標準入力の読込 (`ReadInput`)、ファイル/標準出力への書込 (`WriteOutput`) | - |
| **`envutil`** | **環境変数**の取得と型変換を安全に行うヘルパーを提供します。 | 環境変数取得 (`GetEnv`)、ブール値への変換 (`GetEnvAsBool`) | - |
| **`timeutil`** | **日本標準時 (JST) への変換**など、時刻処理を単純化します。 | JST現在時刻の取得 (`NowJST`)、任意の時刻をJSTへ変換 (`ToJST`) | - |
| **`retry`** | 一時的なエラーに対応するための、**指数バックオフリトライ**を提供します。 | バックオフ付きリトライ実行 (`Do`) | `cenkalti/backoff` 利用 |
| **`text`** | テキストデータのクリーンアップと整形を行います。 | 絵文字除去 (`CleanStringFromEmojis`)、マルチバイト対応の切詰め (`Truncate`)、リストパース | `forPelevin/gomoji` 利用 |

---

## 🚀 クイックスタート

### URLの安全性チェック (`security`)

```go
import "github.com/shouni/go-utils/security"

safe, err := security.IsSafeURL("https://example.com")
if safe && err == nil {
    // 安全なURLに対する処理
}

```

### パスの解決 (`urlpath`)

```go
import "github.com/shouni/go-utils/urlpath"

// リモート(gs://等)かローカルかを問わず、適切にパスを結合します
path, _ := urlpath.ResolvePath("gs://my-bucket/images", "photo.png")
// path => "gs://my-bucket/images/photo.png"

```

---

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

---
