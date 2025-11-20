package urlpath

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cleanURLRegex はファイルシステムやGCSで使用できない文字を特定するための正規表現です。
// \w (単語文字 [a-zA-Z0-9_]) と - (ハイフン) 以外を全てマッチさせます。
var cleanURLRegex = regexp.MustCompile(`[^\w\-]`)

// consecutiveHyphensRegex は連続するハイフンを検出するための正規表現です。
var consecutiveHyphensRegex = regexp.MustCompile(`-{2,}`)

// generateSafeUniqueName は、URLからサニタイズされた安全で一意なディレクトリ名またはGCSキー名を生成する、
// プライベートなヘルパー関数です。
func generateSafeUniqueName(repoURL string) string {
	// 1. net/url を使用してURLをパースし、スキーム、ユーザー情報などを正確に除去
	u, err := url.Parse(repoURL)
	var rawName string

	if err == nil && u.Host != "" {
		// パースが成功した場合 (例: https://github.com/user/repo)
		// ホスト名 + パス を取得
		rawName = u.Host + u.Path
	} else if strings.HasPrefix(repoURL, "git@") {
		// パースが失敗するか、URLライブラリがうまく扱えないGitのSSH形式 (例: git@github.com:user/repo.git)
		// 'git@' を取り除き、残りを rawName として使用
		rawName = strings.TrimPrefix(repoURL, "git@")
		// ホストとパスを分離するコロンをスラッシュに置換 (例: github.com:user/repo -> github.com/user/repo)
		rawName = strings.ReplaceAll(rawName, ":", "/")
	} else {
		// それ以外の形式 (パースエラー、あるいは不明な形式)
		rawName = repoURL
	}

	// rawName の末尾から .git を取り除く (SSH形式の処理もこれでカバー)
	name := strings.TrimSuffix(rawName, ".git")

	// 2. パスとして使用できない文字をハイフンに置換
	name = cleanURLRegex.ReplaceAllString(name, "-")

	// 3. 連続するハイフンを一つにまとめる
	name = consecutiveHyphensRegex.ReplaceAllString(name, "-")
	name = strings.TrimPrefix(name, "-") // 先頭のハイフンを削除
	name = strings.TrimSuffix(name, "-") // 末尾のハイフンを削除

	// 4. 衝突防止のため、URL全体のSHA-256ハッシュの先頭8桁を追加
	hasher := sha256.New()
	hasher.Write([]byte(repoURL))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

	// nameが空の場合はハッシュのみを使用し、不正なパス名を防ぐ
	var safeDirName string
	if name != "" {
		safeDirName = fmt.Sprintf("%s-%s", name, hash)
	} else {
		safeDirName = hash
	}

	return safeDirName
}

// SanitizeURLToUniquePath は、URL をサニタイズ（清浄化）して、ローカルファイルシステム上の
// 一意な一時ディレクトリのフルパスを生成します。
func SanitizeURLToUniquePath(repoURL string, baseRepoDirName string) string {
	// 1. ユニークな名前部分をヘルパー関数で生成
	safeDirName := generateSafeUniqueName(repoURL)

	// 2. OS互換のため filepath.Join を使用し、一時ディレクトリと結合して返す
	tempBase := filepath.Join(os.TempDir(), baseRepoDirName)
	return filepath.Join(tempBase, safeDirName)
}

// GenerateGCSKeyName は、リポジトリURLからGCSオブジェクトキーの一部として
// 使用するための、安全で一意なディレクトリ名（ローカルパスではない）を生成します。
func GenerateGCSKeyName(repoURL string) string {
	// ヘルパー関数を呼び出し、GCSキー名として必要なユニーク名のみを返す
	return generateSafeUniqueName(repoURL)
}
