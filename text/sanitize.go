package text

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forPelevin/gomoji"
)

// cleanURLRegex はファイルシステムで使用できない文字を特定するための正規表現です。
var cleanURLRegex = regexp.MustCompile(`[^\w\-.]+`)

// NormalizeText は、連続する空白文字（改行やタブを含む）を単一のスペースに変換し、
// 前後の空白を削除します。内部で strings.Fields を使用しています。
func NormalizeText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// RemoveEmojis は、入力文字列からすべての絵文字文字を削除します。
func RemoveEmojis(s string) string {
	return gomoji.RemoveEmojis(s)
}

// CleanStringFromEmojis は、入力文字列からすべての絵文字文字を削除し、空白を正規化します。
// この関数は、最初に絵文字を削除し、次に NormalizeText を使用して空白を正規化します。
func CleanStringFromEmojis(s string) string {
	s = RemoveEmojis(s)
	s = NormalizeText(s)
	return s
}

// Truncate は、指定された文字列を rune (文字) の数で最大長まで切り詰め、必要に応じてサフィックスを追加します。
//
// 注意: 切り詰められた文字列の末尾に空白文字が残った場合、サフィックスを付加する前に
// strings.TrimSpaceを使用してその末尾の空白は無条件に削除されます。
// これは、ログ出力や表示用途で不要な末尾スペースを排除し、サフィックスを綺麗に付加するための仕様です
func Truncate(s string, maxLen int, suffix string) string {
	if maxLen <= 0 {
		return ""
	}

	// 1. 文字列を rune のスライスに変換 (マルチバイト対応)
	runes := []rune(s)

	// 2. 文字数が最大長以下であればそのまま返す
	if len(runes) <= maxLen {
		return s
	}

	// 3. rune の数で切り詰める
	truncatedRuneSlice := runes[:maxLen]

	// 4. rune スライスを文字列に戻し、末尾スペースを削除
	truncatedString := strings.TrimSpace(string(truncatedRuneSlice))

	return truncatedString + suffix
}

// SanitizeURLToUniquePath は、URL をサニタイズ（清浄化）して、一意なパス文字列を生成します。
func SanitizeURLToUniquePath(repoURL string) string {
	// ベースディレクトリを設定 (例: /tmp/git-reviewer-repos)
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
