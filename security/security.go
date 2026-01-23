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

	scheme := strings.ToLower(u.Scheme)
	hostname := strings.ToLower(u.Hostname())

	switch scheme {
	case SchemeHTTPS:
		return true
	case SchemeHTTP:
		return isLocalDevHostname(hostname)
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

	scheme := strings.ToLower(parsedURL.Scheme)

	// クラウドストレージ用スキームは信頼済みとして早期リターン
	if scheme == SchemeGCS || scheme == SchemeS3 {
		return true, nil
	}

	if scheme != SchemeHTTP && scheme != SchemeHTTPS {
		return false, fmt.Errorf("不許可スキーム: %s", parsedURL.Scheme)
	}

	hostname := strings.ToLower(parsedURL.Hostname())
	if hostname == "" {
		return false, fmt.Errorf("ホストが空です")
	}

	if err := validateHostnameIPs(hostname); err != nil {
		return false, err
	}

	return true, nil
}

// isLocalDevHostname 指定されたホスト名が信頼できるローカル開発ホスト名であるかどうかを確認します。
func isLocalDevHostname(hostname string) bool {
	if hostname == "" {
		return false
	}
	_, ok := localdevHostnames[hostname]
	return ok
}

// validateHostnameIPs 指定されたホスト名の IP を解決し、制限されたネットワーク範囲へのアクセスをチェックします。
// DNS 解決に失敗した場合、または解決された IP が制限された範囲に属している場合はエラーを返します。
func validateHostnameIPs(hostname string) error {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("ホスト '%s' の名前解決に失敗: %w", hostname, err)
	}

	for _, ip := range ips {
		if isRestrictedIP(ip) {
			return fmt.Errorf("制限されたネットワークへのアクセスを検知: %s", ip.String())
		}
	}
	return nil
}

// isRestrictedIP は、IPがプライベート、ループバック、またはリンクローカルであるか判定します。
func isRestrictedIP(ip net.IP) bool {
	return ip.IsPrivate() ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast()
}
