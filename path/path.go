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

// SanitizeURLToUniquePath は、URL をサニタイズ（清浄化）して、一意なパス文字列を生成します。
func SanitizeURLToUniquePath(repoURL string) string {
	// ベースディレクトリを設定
	tempBase := os.TempDir() + "/git-reviewer-repos"

	// 1. スキームと.gitを削除してクリーンな名前を取得
	name := strings.TrimSuffix(repoURL, ".git")
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "git@")
	name = strings.Trim(name, "-")

	// 2. パスとして使用できない文字をハイフンに置換
	// cleanURLRegex を使用して、ファイルシステムで安全でない文字を置換
	name = cleanURLRegex.ReplaceAllString(name, "-")

	// 3. 連続するハイフンを一つにまとめる
	// これにより、スキームやパス区切り文字が変換された結果の連続ハイフンがクリーンになる
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")

	// 4. 衝突防止のため、URL全体のSHA-256ハッシュの先頭8桁を追加
	hasher := sha256.New()
	hasher.Write([]byte(repoURL))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

	// パス名が長くなりすぎないように調整し、ハイフンをトリム
	safeDirName := fmt.Sprintf("%s-%s", name, hash)

	// 5. ベースパスと結合して返す
	return filepath.Join(tempBase, safeDirName)
}
