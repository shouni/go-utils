# 📚 Go Utils

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`Go Utils`** は、Go言語でアプリケーションを開発する際に繰り返し必要となる、汎用的で再利用可能なユーティリティ機能を集めたライブラリです。

このプロジェクトは、それぞれの機能が独立したパッケージとして提供されており、必要な機能のみをアプリケーションにインポートして利用することで、クリーンな依存関係を維持できます。

## ✨ 特徴

* **モジュール性**: 各ユーティリティが独立したGoパッケージとして提供されます。
* **汎用性**: 特定のアプリケーションドメインに依存しない、幅広いタスクに適用可能です。
* **信頼性**: 外部ライブラリ（例: `cenkalti/backoff`）を利用して、堅牢な機能を提供します。

-----

## 🛠️ インストール

プロジェクト全体をインストールするには、以下のコマンドを使用します。

```bash
go get github.com/shouni/go-utils
```

特定のパッケージのみを利用する場合は、そのパッケージをインポートしてください（例: `import "github.com/shouni/go-utils/timeutil"`）。

-----

## 📦 パッケージ構成 (Package Structure)

以下のパッケージがこのリポジトリで提供されています。

| パッケージ | 説明 | 主な提供機能 | 関連情報 |
| :--- | :--- | :--- | :--- |
| **`iohandler`** | **ファイルI/Oと標準入出力（stdin/stdout）を抽象化**します。CLIアプリケーションでの入出力処理を簡潔にします。 | ファイルまたは標準入力からのコンテンツの**読み込み** (`ReadInput`)、ファイルまたは標準出力へのコンテンツの**書き込み** (`WriteOutput`) | - |
| **`envutil`** | **環境変数**を取得し、存在しない場合にデフォルト値を適用したり、**型変換**を行ったりするヘルパーを提供します。 | 環境変数取得 (`GetEnv`)、型変換付き環境変数取得 (`GetEnvAsBool` など) | **新規追加** |
| **`urlpath`** | **URLの解析**を行い、ファイルシステムパスへの変換や、**Webサービス設定におけるURLスキームの安全性の判定**に使用します。 | URLのサニタイズ、一意なキャッシュパスの生成 (`SanitizeURLToUniquePath`, `GenerateGCSKeyName`)、**Webサービスのセキュア判定** (`IsSecureServiceURL`) | - |
| **`timeutil`** | **日本標準時 (JST) の取得と変換**など、時刻とタイムゾーン処理を単純化します。 | JSTの現在時刻の取得 (`NowJST`)、任意の時刻をJSTへ変換 (`ToJST`) | **新規追加** |
| **`retry`** | 外部サービス連携などで発生する**一時的なエラーに対応**するための、汎用的なリトライロジックを提供します。 | 指数バックオフ、カスタムエラー判定 (`Do` 関数) | `github.com/cenkalti/backoff/v4` を利用 |
| **`text`** | テキストデータのクリーンアップと整形を行います。特に、**非互換な文字の除去**と**表示の切り詰め**に役立ちます。 | **絵文字の除去**と空白の正規化 (`CleanStringFromEmojis`)、**マルチバイト対応の文字列切り詰め** (`Truncate`)、**カンマ区切りのリストパース** (`ParseCommaSeparatedList`) | `github.com/forPelevin/gomoji` を利用 |
| （その他） | 必要に応じて追加されるその他のユーティリティパッケージ。 | ロギング、文字列操作、設定管理など | |

-----

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。
