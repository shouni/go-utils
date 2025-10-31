# 📚 Go Utils

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-utils)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-utils)](https://github.com/shouni/go-utils/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`go-utils`** は、Go言語でアプリケーションを開発する際に繰り返し必要となる、汎用的で再利用可能なユーティリティ機能を集めたライブラリです。

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

特定のパッケージのみを利用する場合は、そのパッケージをインポートしてください（例: `import "github.com/shouni/go-utils/retry"`）。

-----

## 📦 パッケージ構成 (Package Structure)

以下のパッケージがこのリポジトリで提供されています。

| パッケージ | 説明 | 主な提供機能 | 関連情報 |
| :--- | :--- | :--- | :--- |
| **`retry`** | 外部サービス連携などで発生する**一時的なエラーに対応**するための、汎用的なリトライロジックを提供します。 | 指数バックオフ、カスタムエラー判定 (`Do` 関数) | `github.com/cenkalti/backoff/v4` を利用 |
| **`text`** | テキストデータのクリーンアップと整形を行います。特に、**非互換な文字の除去**に役立ちます。 | **絵文字の除去**と、それに伴う**厳密な空白の正規化** (`CleanStringFromEmojis`) | `github.com/forPelevin/gomoji` を利用 |
| （その他） | 必要に応じて追加されるその他のユーティリティパッケージ。 | ロギング、文字列操作、設定管理など | |

-----

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

