package path

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cleanURLRegex はファイルシステムで使用できない文字を特定するための正規表現です。
var cleanURLRegex = regexp.MustCompile(`[^\w\-.]+`)

// consecutiveHyphensRegex は連続するハイフンを検出するための正規表現です。
// (修正案3: パフォーマンス向上のためグローバル変数としてコンパイル)
var consecutiveHyphensRegex = regexp.MustCompile(`-+`)

// baseRepoDirName はテンポラリディレクトリ内に作成されるリポジトリキャッシュディレクトリの基本名です。
// (修正案1: マジックストリングの定数化)
const baseRepoDirName = "git-reviewer-repos"

// SanitizeURLToUniquePath は、URL をサニタイズ（清浄化）して、一意なパス文字列を生成します。
func SanitizeURLToUniquePath(repoURL string) string {
	// (修正案1: OS互換のため filepath.Join を使用)
	tempBase := filepath.Join(os.TempDir(), baseRepoDirName)

	// 1. スキームと.gitを削除してクリーンな名前を取得
	name := strings.TrimSuffix(repoURL, ".git")
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "git@")

	// 2. パスとして使用できない文字をハイフンに置換
	name = cleanURLRegex.ReplaceAllString(name, "-")

	// 3. 連続するハイフンを一つにまとめる
	// (修正案3: グローバル変数を使用)
	name = consecutiveHyphensRegex.ReplaceAllString(name, "-")

	// (修正案2: strings.Trim の代わりに、意図が明確な TrimPrefix/Suffix を使用)
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimSuffix(name, "-")

	// 4. 衝突防止のため、URL全体のSHA-256ハッシュの先頭8桁を追加
	hasher := sha256.New()
	hasher.Write([]byte(repoURL))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

	// (修正案4: nameが空の場合のフォールバックロジック)
	// nameが空の場合はハッシュのみを使用し、"-<hash>" のような不正なパス名を防ぐ
	var safeDirName string
	if name != "" {
		safeDirName = fmt.Sprintf("%s-%s", name, hash)
	} else {
		safeDirName = hash
	}

	// 5. ベースパスと結合して返す
	return filepath.Join(tempBase, safeDirName)
}
