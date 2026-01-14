package urlpath_test

import (
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
