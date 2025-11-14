package urlpath_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shouni/go-utils/urlpath"
)

// NOTE: このテストでは、ハッシュ部分の検証は行わず、
// パスが適切な構造と整形された名前を持っていることを検証します。

func TestSanitizeURLToUniquePath(t *testing.T) {
	tempBase := filepath.Join(os.TempDir(), "reviewer-repos")

	tests := []struct {
		name     string
		inputURL string
		// 期待される整形された名前のプレフィックス (ハッシュと結合される部分)
		expectedPrefix string
		// URLをクリーンアップした後に生成される最終的なパスのベース部分
		expectedPathBase string
	}{
		{
			name:             "HTTP_Basic",
			inputURL:         "https://github.com/owner/repo.git",
			expectedPrefix:   "github-com-owner-repo",
			expectedPathBase: tempBase,
		},
		{
			name:             "HTTPS_NoGitSuffix",
			inputURL:         "https://gitlab.com/group/subgroup/project",
			expectedPrefix:   "gitlab-com-group-subgroup-project",
			expectedPathBase: tempBase,
		},
		{
			name:     "SSH_Protocol",
			inputURL: "git@bitbucket.org:team/project.git",
			// git@ と : がハイフンに変換され、連続ハイフンが除去される
			expectedPrefix:   "bitbucket-org-team-project",
			expectedPathBase: tempBase,
		},
		{
			name:     "TrailingSlash",
			inputURL: "https://example.com/project/",
			// 末尾のスラッシュは cleanURLRegex でハイフンになり、TrimSuffixで除去される
			expectedPrefix:   "example-com-project",
			expectedPathBase: tempBase,
		},
		{
			name:     "EmptyURL",
			inputURL: "",
			// nameが空になり、ハッシュのみがパス名になる
			expectedPrefix:   "",
			expectedPathBase: tempBase,
		},
		{
			name:             "OnlyScheme",
			inputURL:         "http://",
			expectedPrefix:   "",
			expectedPathBase: tempBase,
		},
		{
			name:     "LongURL",
			inputURL: "https://long.domain.name/with/many/path/segments/to/test/if/it/handles/long/strings/and/converts/all/the/slashes/and/dots/correctly/and/trims/the/prefix.git",
			// 連続ハイフン処理と、適切な Trim が行われることを確認する
			// ここでは完全な prefix の代わりに、整形された一部をチェックする
			expectedPrefix:   "long-domain-name-with-many-path-segments-to-test-if-it-handles-long-strings-and-converts-all-the-slashes-and-dots-correctly-and-trims-the-prefix",
			expectedPathBase: tempBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultPath := urlpath.SanitizeURLToUniquePath(tt.inputURL)

			// 1. ベースパスが期待通りか検証
			if !strings.HasPrefix(resultPath, tt.expectedPathBase) {
				t.Errorf("SanitizeURLToUniquePath(%q) = %q, expected path to start with %q", tt.inputURL, resultPath, tt.expectedPathBase)
				return
			}

			// 2. ディレクトリ名部分を抽出
			// ディレクトリ名は tempBase の後の部分
			dirName := strings.TrimPrefix(resultPath, tt.expectedPathBase+string(filepath.Separator))

			// 3. ハッシュ部分を検証 (長さとフォーマット)
			parts := strings.Split(dirName, "-")
			hashPart := parts[len(parts)-1]

			if len(hashPart) != 8 {
				t.Errorf("SanitizeURLToUniquePath(%q) hash length incorrect: got %d, want 8", tt.inputURL, len(hashPart))
			}

			// 4. 整形された名前部分を検証 (ハッシュ部分を除く)
			prefixPart := strings.Join(parts[:len(parts)-1], "-")

			if tt.inputURL == "" || tt.inputURL == "http://" {
				// EmptyURL または OnlyScheme の場合、nameは空で、prefixPartは空になる
				if prefixPart != "" {
					t.Errorf("SanitizeURLToUniquePath(%q) prefix incorrect for empty input: got %q, want \"\"", tt.inputURL, prefixPart)
				}
			} else if prefixPart != tt.expectedPrefix {
				t.Errorf("SanitizeURLToUniquePath(%q) prefix incorrect.\nGot:  %q\nWant: %q", tt.inputURL, prefixPart, tt.expectedPrefix)
			}

			// 5. パスに連続ハイフンや先頭/末尾のハイフンが残っていないことを確認 (目視チェック)
			if strings.Contains(prefixPart, "--") || strings.HasPrefix(prefixPart, "-") || strings.HasSuffix(prefixPart, "-") {
				t.Errorf("SanitizeURLToUniquePath(%q) contains consecutive/leading/trailing hyphens: %q", tt.inputURL, prefixPart)
			}
		})
	}
}
