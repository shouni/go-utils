package urlpath_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shouni/go-utils/urlpath"
)

// ----------------------------------------------------------------------
// TestIsSecureServiceURL: セキュリティ検証関数のテスト
// ----------------------------------------------------------------------

func TestIsSecureServiceURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		want     bool
	}{
		{
			name:     "HTTPS_Secure",
			inputURL: "https://example.com/api",
			want:     true,
		},
		{
			name:     "HTTP_Insecure",
			inputURL: "http://production.com/api",
			want:     false,
		},
		{
			name:     "HTTP_Localhost_Secure",
			inputURL: "http://localhost:8080/api",
			want:     true, // 許可されたローカル開発環境
		},
		{
			name:     "HTTP_127001_Secure",
			inputURL: "http://127.0.0.1/auth",
			want:     true, // 許可されたローカル開発環境
		},
		{
			name:     "HTTP_IPv6Local_Secure",
			inputURL: "http://[::1]:3000",
			want:     true, // 許可されたローカル開発環境
		},
		{
			name:     "HTTP_WithCapitalHost_Secure",
			inputURL: "http://LocalHost:8080/path",
			want:     true, // ホスト名を小文字に変換してチェックされる
		},
		{
			name:     "UnknownScheme_Insecure",
			inputURL: "ftp://fileserver.net/data",
			want:     false,
		},
		{
			name:     "InvalidURL_Insecure",
			inputURL: "::invalid-url",
			want:     false, // パースエラーで false
		},
		{
			name:     "EmptyURL_Insecure",
			inputURL: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlpath.IsSecureServiceURL(tt.inputURL)
			if got != tt.want {
				t.Errorf("IsSecureServiceURL(%q) = %v, want %v", tt.inputURL, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------
// TestGetRepositoryPath: リポジトリパス抽出関数のテスト (新規追加)
// ----------------------------------------------------------------------

func TestGetRepositoryPath(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		want     string
	}{
		{
			name:     "HTTPS_GitSuffix",
			inputURL: "https://github.com/shouni/go-utils.git",
			want:     "shouni/go-utils",
		},
		{
			name:     "HTTPS_NoGitSuffix",
			inputURL: "https://gitlab.com/group/project",
			want:     "group/project",
		},
		{
			name:     "SSH_Protocol_Colon",
			inputURL: "git@bitbucket.org:team/repository.git",
			want:     "team/repository", // git@ と .git が除去され、: が / に変換される
		},
		{
			name:     "SSH_Protocol_LongPath",
			inputURL: "git@host.net:user/subgroup/repo-name.git",
			want:     "user/subgroup/repo-name",
		},
		{
			name:     "SSH_URLScheme",
			inputURL: "ssh://git@github.com/owner/repo.git",
			want:     "owner/repo", // net/url が処理
		},
		{
			name:     "URL_SubDir",
			inputURL: "https://example.com/owner/repo/subdir",
			want:     "owner/repo/subdir",
		},
		{
			name:     "OnlyHostAndOwner",
			inputURL: "https://github.com/owner",
			want:     "owner",
		},
		{
			name:     "EmptyURL",
			inputURL: "",
			want:     "", // パースエラー時、元のURLがそのまま返される
		},
		{
			name:     "InvalidURL",
			inputURL: "::invalid-url",
			want:     "::invalid-url", // パースエラー時、元のURLがそのまま返される
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlpath.GetRepositoryPath(tt.inputURL)
			if got != tt.want {
				t.Errorf("GetRepositoryPath(%q) = %q, want %q", tt.inputURL, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------
// TestGenerateGCSKeyName: GCSキー生成関数のテスト
// ----------------------------------------------------------------------

func TestGenerateGCSKeyName(t *testing.T) {
	// GenerateGCSKeyNameは SanitizeURLToUniquePath の名前生成ロジックと同一
	// (generateSafeUniqueName を呼び出す) のため、SanitizeURLToUniquePath のテストケースを再利用
	tests := []struct {
		name           string
		inputURL       string
		expectedPrefix string
	}{
		{
			name:           "HTTP_Basic",
			inputURL:       "https://github.com/owner/repo.git",
			expectedPrefix: "github-com-owner-repo",
		},
		{
			name:           "SSH_Protocol_Colon",
			inputURL:       "git@bitbucket.org:team/project.git",
			expectedPrefix: "bitbucket-org-team-project",
		},
		{
			name:           "EmptyURL",
			inputURL:       "",
			expectedPrefix: "", // 名前部分は空になり、ハッシュのみになる
		},
		{
			name:           "URLWithSpecialChars",
			inputURL:       "https://test.com/project_name-with.dots-and_underscores",
			expectedPrefix: "test-com-project_name-with-dots-and_underscores", // ドットはハイフンに、アンダースコアはそのまま
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := urlpath.GenerateGCSKeyName(tt.inputURL)

			// 1. ハッシュ部分の検証 (長さとフォーマット)
			parts := strings.Split(result, "-")
			hashPart := parts[len(parts)-1]

			if len(hashPart) != 8 {
				t.Errorf("GenerateGCSKeyName(%q) hash length incorrect: got %d, want 8", tt.inputURL, len(hashPart))
			}

			// 2. 整形された名前部分を検証 (ハッシュ部分を除く)
			prefixPart := strings.Join(parts[:len(parts)-1], "-")

			if tt.inputURL == "" {
				if prefixPart != "" {
					t.Errorf("GenerateGCSKeyName(%q) prefix incorrect for empty input: got %q, want \"\"", tt.inputURL, prefixPart)
				}
			} else if prefixPart != tt.expectedPrefix {
				t.Errorf("GenerateGCSKeyName(%q) prefix incorrect.\nGot:  %q\nWant: %q", tt.inputURL, prefixPart, tt.expectedPrefix)
			}
		})
	}
}

// ----------------------------------------------------------------------
// TestSanitizeURLToUniquePath: ローカルパス生成関数のテスト
// ----------------------------------------------------------------------

// NOTE: このテストでは、ハッシュ部分の検証は行わず、
// パスが適切な構造と整形された名前を持っていることを検証します。

func TestSanitizeURLToUniquePath(t *testing.T) {
	// 💡 baseRepoDirName を定義
	const baseDirName = "reviewer-repos"
	tempBase := filepath.Join(os.TempDir(), baseDirName)

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
			name:     "SSH_Protocol_Colon",
			inputURL: "git@bitbucket.org:team/project.git",
			// git@ は TrimPrefix、: は / に変換され、/ は cleanURLRegex でハイフンになる
			expectedPrefix:   "bitbucket-org-team-project",
			expectedPathBase: tempBase,
		},
		{
			name:     "SSH_Protocol_URLScheme",
			inputURL: "ssh://git@github.com/owner/repo.git",
			// ssh:// も net/url がスキームとして除去する
			expectedPrefix:   "github-com-owner-repo",
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
			name:     "OnlyScheme",
			inputURL: "http://",
			// net/url でホストが空になり、rawNameが "http://" のまま残るため、cleanURLRegexで "http--" になり、
			// 連続ハイフン処理で "http" になる
			expectedPrefix:   "http",
			expectedPathBase: tempBase,
		},
		{
			name:             "LongURL",
			inputURL:         "https://long.domain.name/with/many/path/segments/to/test/if/it/handles/long/strings/and/converts/all/the/slashes/and/dots/correctly/and/trims/the/prefix.git",
			expectedPrefix:   "long-domain-name-with-many-path-segments-to-test-if-it-handles-long-strings-and-converts-all-the-slashes-and-dots-correctly-and-trims-the-prefix",
			expectedPathBase: tempBase,
		},
		{
			name:             "URLWithPort",
			inputURL:         "https://dev.example.com:8080/repo",
			expectedPrefix:   "dev-example-com-repo", // ポート番号は net/url により適切に除去される
			expectedPathBase: tempBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// baseDirName を第2引数として渡すように修正
			resultPath := urlpath.SanitizeURLToUniquePath(tt.inputURL, baseDirName)

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

			// OnlyScheme の expectedPrefix 修正に伴い、条件を調整
			// EmptyURL の場合のみ name は空
			if tt.inputURL == "" {
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
