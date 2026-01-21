package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// 安全なスキームの定義
const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
	SchemeGCS   = "gs"
	SchemeS3    = "s3"
)

// localdevHostnames は、HTTPでも安全と見なすホスト名のセットです。
var localdevHostnames = map[string]struct{}{
	"localhost":            {},
	"127.0.0.1":            {},
	"::1":                  {},
	"host.docker.internal": {}, // Docker環境用に追加
}

// IsSecureServiceURL は、URLがHTTPSであるか、信頼できるローカル環境であるかを確認します。
func IsSecureServiceURL(serviceURL string) bool {
	u, err := url.Parse(serviceURL)
	if err != nil {
		return false
	}

	switch u.Scheme {
	case SchemeHTTPS:
		return true
	case SchemeHTTP:
		hostname := strings.ToLower(u.Hostname())
		_, isLocal := localdevHostnames[hostname]
		return isLocal
	default:
		return false
	}
}

// IsSafeURL は、SSRF対策としてURLを検証します。
func IsSafeURL(rawURL string) (bool, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false, fmt.Errorf("URLパース失敗: %w", err)
	}

	// クラウドストレージ用スキームは信頼済みとして早期リターン
	switch parsedURL.Scheme {
	case SchemeGCS, SchemeS3:
		return true, nil
	case SchemeHTTP, SchemeHTTPS:
		// 続行してIP検証へ
	default:
		return false, fmt.Errorf("不許可スキーム: %s", parsedURL.Scheme)
	}

	// ホスト名からIPアドレスを取得して検証
	hostname := parsedURL.Hostname()
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return false, fmt.Errorf("ホスト '%s' の名前解決に失敗: %w", hostname, err)
	}

	for _, ip := range ips {
		if isRestrictedIP(ip) {
			return false, fmt.Errorf("制限されたネットワークへのアクセスを検知: %s", ip.String())
		}
	}

	return true, nil
}

// isRestrictedIP は、IPがプライベート、ループバック、またはリンクローカルであるか判定します。
func isRestrictedIP(ip net.IP) bool {
	return ip.IsPrivate() ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast()
}
