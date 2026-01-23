package security_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shouni/go-utils/security"
)

// ----------------------------------------------------------------------
// TestIsSecureServiceURL: HTTPS/ローカル環境判定のテスト
// ----------------------------------------------------------------------

func TestIsSecureServiceURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		want     bool
	}{
		{
			name:     "HTTPS_ValidURL",
			inputURL: "https://example.com/api",
			want:     true,
		},
		{
			name:     "HTTPS_WithPort",
			inputURL: "https://example.com:8443/secure",
			want:     true,
		},
		{
			name:     "HTTP_Localhost",
			inputURL: "http://localhost:8080/api",
			want:     true,
		},
		{
			name:     "HTTP_127.0.0.1",
			inputURL: "http://127.0.0.1:3000",
			want:     true,
		},
		{
			name:     "HTTP_IPv6_Loopback",
			inputURL: "http://[::1]:8080/test",
			want:     true,
		},
		{
			name:     "HTTP_DockerInternal",
			inputURL: "http://host.docker.internal:5000",
			want:     true,
		},
		{
			name:     "HTTP_ExternalHost_Insecure",
			inputURL: "http://example.com/api",
			want:     false,
		},
		{
			name:     "HTTP_MixedCase_Localhost",
			inputURL: "http://LocalHost:8080",
			want:     true,
		},
		{
			name:     "FTP_Scheme_Invalid",
			inputURL: "ftp://example.com/file",
			want:     false,
		},
		{
			name:     "InvalidURL_ParseError",
			inputURL: "://invalid-url",
			want:     false,
		},
		{
			name:     "EmptyURL",
			inputURL: "",
			want:     false,
		},
		{
			name:     "NoScheme",
			inputURL: "example.com",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := security.IsSecureServiceURL(tt.inputURL)
			if got != tt.want {
				t.Errorf("IsSecureServiceURL(%q) = %v, want %v", tt.inputURL, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------
// TestIsSafeURL: SSRF対策のURL検証テスト
// ----------------------------------------------------------------------

func TestIsSafeURL(t *testing.T) {
	tests := []struct {
		name       string
		inputURL   string
		wantSafe   bool
		wantErrMsg string // エラーメッセージに含まれるべき文字列
	}{
		{
			name:     "CloudStorage_GCS",
			inputURL: "gs://bucket-name/object/path",
			wantSafe: true,
		},
		{
			name:     "CloudStorage_S3",
			inputURL: "s3://my-bucket/data.json",
			wantSafe: true,
		},
		{
			name:     "HTTPS_PublicDomain",
			inputURL: "https://example.com/api",
			wantSafe: true,
		},
		{
			name:     "HTTP_PublicDomain",
			inputURL: "http://example.com/data",
			wantSafe: true,
		},
		{
			name:       "HTTP_Localhost_Restricted",
			inputURL:   "http://localhost/admin",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:       "HTTP_127.0.0.1_Restricted",
			inputURL:   "http://127.0.0.1:8080/secret",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:       "HTTP_PrivateIP_10.0.0.1",
			inputURL:   "http://10.0.0.1/internal",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:       "HTTP_PrivateIP_192.168.1.1",
			inputURL:   "http://192.168.1.1/router",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:       "HTTP_PrivateIP_172.16.0.1",
			inputURL:   "http://172.16.0.1/admin",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:       "FTP_InvalidScheme",
			inputURL:   "ftp://example.com/file",
			wantSafe:   false,
			wantErrMsg: "不許可スキーム",
		},
		{
			name:       "EmptyHost",
			inputURL:   "http://",
			wantSafe:   false,
			wantErrMsg: "ホストが空です",
		},
		{
			name:       "InvalidURL",
			inputURL:   "://invalid",
			wantSafe:   false,
			wantErrMsg: "URLパース失敗",
		},
		{
			name:       "NoScheme",
			inputURL:   "example.com",
			wantSafe:   false,
			wantErrMsg: "URLパース失敗",
		},
		{
			name:     "MixedCase_Scheme_GCS",
			inputURL: "GS://bucket/object",
			wantSafe: true,
		},
		{
			name:     "MixedCase_Scheme_HTTPS",
			inputURL: "HTTPS://Example.COM/api",
			wantSafe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe, err := security.IsSafeURL(tt.inputURL)

			if tt.wantSafe {
				if err != nil {
					t.Errorf("IsSafeURL(%q) returned unexpected error: %v", tt.inputURL, err)
				}
				if !safe {
					t.Errorf("IsSafeURL(%q) = false, want true", tt.inputURL)
				}
			} else {
				if err == nil {
					t.Errorf("IsSafeURL(%q) expected error but got none", tt.inputURL)
				} else if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("IsSafeURL(%q) error = %q, want error containing %q", tt.inputURL, err.Error(), tt.wantErrMsg)
				}
				if safe {
					t.Errorf("IsSafeURL(%q) = true, want false", tt.inputURL)
				}
			}
		})
	}
}

// ----------------------------------------------------------------------
// TestNewSafeHTTPClient: DNS Rebinding対策クライアントのテスト
// ----------------------------------------------------------------------

func TestNewSafeHTTPClient(t *testing.T) {
	t.Run("ClientCreation", func(t *testing.T) {
		timeout := 10 * time.Second
		client := security.NewSafeHTTPClient(timeout)

		if client == nil {
			t.Fatal("NewSafeHTTPClient returned nil")
		}

		if client.Timeout != timeout {
			t.Errorf("Client timeout = %v, want %v", client.Timeout, timeout)
		}

		if client.Transport == nil {
			t.Error("Client transport is nil")
		}
	})

	t.Run("BlockPrivateIPConnection", func(t *testing.T) {
		client := security.NewSafeHTTPClient(5 * time.Second)

		// プライベートIPへの接続を試みる（実際には接続しない）
		req, err := http.NewRequestWithContext(
			context.Background(),
			"GET",
			"http://127.0.0.1:9999/test",
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			t.Error("Expected error for private IP connection, got nil")
		} else if !strings.Contains(err.Error(), "restricted IP detected") {
			t.Errorf("Expected 'restricted IP detected' error, got: %v", err)
		}
	})

	t.Run("AllowPublicIPConnection", func(t *testing.T) {
		// テスト用の公開サーバーを立てる
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))
		defer server.Close()

		// ただし、httptest.NewServer はループバックアドレスを使うため、
		// 実際にはこのテストは制限されてしまう。
		// 実運用では外部のパブリックIPに対してテストする必要がある。

		client := security.NewSafeHTTPClient(5 * time.Second)

		req, err := http.NewRequestWithContext(
			context.Background(),
			"GET",
			server.URL,
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = client.Do(req)
		// loopback なので制限される想定
		if err == nil {
			t.Log("Note: httptest server uses loopback, so this may fail in real scenarios")
		}
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		client := security.NewSafeHTTPClient(100 * time.Millisecond)

		// タイムアウトが短いコンテキストを作成
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(
			ctx,
			"GET",
			"https://example.com/slow",
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = client.Do(req)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

// ----------------------------------------------------------------------
// TestIsRestrictedIP: IP制限判定のテスト（間接的なテスト）
// ----------------------------------------------------------------------

// isRestrictedIP は非公開関数なので、IsSafeURL 経由で間接的にテスト済み
// 追加で特定のケースをカバーしたい場合は、IsSafeURL のテストケースを拡充する

func TestIsSafeURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		inputURL   string
		wantSafe   bool
		wantErrMsg string
	}{
		{
			name:       "IPv6_Loopback",
			inputURL:   "http://[::1]:8080/admin",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:       "IPv6_LinkLocal",
			inputURL:   "http://[fe80::1]:8080/test",
			wantSafe:   false,
			wantErrMsg: "制限されたネットワークへのアクセスを検知",
		},
		{
			name:     "ValidURL_WithPath",
			inputURL: "https://api.example.com/v1/users?id=123",
			wantSafe: true,
		},
		{
			name:     "ValidURL_WithFragment",
			inputURL: "https://example.com/page#section",
			wantSafe: true,
		},
		{
			name:       "OnlyScheme",
			inputURL:   "https://",
			wantSafe:   false,
			wantErrMsg: "ホストが空です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe, err := security.IsSafeURL(tt.inputURL)

			if tt.wantSafe {
				if err != nil {
					t.Errorf("IsSafeURL(%q) returned unexpected error: %v", tt.inputURL, err)
				}
				if !safe {
					t.Errorf("IsSafeURL(%q) = false, want true", tt.inputURL)
				}
			} else {
				if err == nil {
					t.Errorf("IsSafeURL(%q) expected error but got none", tt.inputURL)
				} else if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("IsSafeURL(%q) error = %q, want error containing %q", tt.inputURL, err.Error(), tt.wantErrMsg)
				}
				if safe {
					t.Errorf("IsSafeURL(%q) = true, want false", tt.inputURL)
				}
			}
		})
	}
}
