package security

import (
	"net"
	"testing"
)

func TestIsSecureServiceURL(t *testing.T) {
	tests := []struct {
		name       string
		serviceURL string
		want       bool
	}{
		{"Valid HTTPS", "https://example.com", true},
		{"Valid HTTP Localhost", "http://localhost/callback", true},
		{"Valid HTTP 127.0.0.1", "http://127.0.0.1:8080", true},
		{"Valid IPv6 Loopback", "http://[::1]", true},
		{"Valid Docker Internal", "http://host.docker.internal:9000", true},
		{"Invalid HTTP Public", "http://example.com", false},
		{"Invalid Scheme FTP", "ftp://localhost", false},
		{"Invalid URL", "::not-a-url::", false},
		{"Empty String", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSecureServiceURL(tt.serviceURL); got != tt.want {
				t.Errorf("IsSecureServiceURL(%q) = %v, want %v", tt.serviceURL, got, tt.want)
			}
		})
	}
}

func TestIsSafeURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    bool
		wantErr bool
	}{
		// クラウドストレージスキーム
		{"Safe GCS", "gs://my-bucket/path", true, false},
		{"Safe S3", "s3://my-bucket/path", true, false},

		// 公開HTTP/HTTPS (名前解決が成功することを期待)
		{"Safe Public HTTPS", "https://www.google.com", true, false},
		{"Safe Public HTTP", "http://example.com", true, false},

		// 不許可スキーム
		{"Unsafe Scheme File", "file:///etc/passwd", false, true},
		{"Unsafe Scheme Metadata", "metadata://google.internal", false, true},

		// SSRF攻撃ベクトル (ループバック/プライベート)
		{"Unsafe Loopback", "http://127.0.0.1", false, true},
		{"Unsafe IPv6 Loopback", "http://[::1]", false, true},
		// 実際の実装では google.com などはOKだが、ローカルIPを指すホストはNG
		// 注: CI環境で名前解決ができないホストを指定するとエラーになります
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsSafeURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSafeURL(%q) error = %v, wantErr %v", tt.rawURL, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsSafeURL(%q) = %v, want %v", tt.rawURL, got, tt.want)
			}
		})
	}
}

func TestIsRestrictedIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"Private IPv4 Class A", "10.0.0.1", true},
		{"Private IPv4 Class B", "172.16.0.1", true},
		{"Private IPv4 Class C", "192.168.1.1", true},
		{"Loopback IPv4", "127.0.0.1", true},
		{"Loopback IPv6", "::1", true},
		{"Link Local", "169.254.169.254", true},
		{"Public IPv4", "8.8.8.8", false},
		{"Public IPv6", "2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRestrictedIP(net.ParseIP(tt.ip)); got != tt.want {
				t.Errorf("isRestrictedIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
