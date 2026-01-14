package urlpath

import (
	"net/url"
	"strings"
)

// localdevHostnames は、HTTPスキームで「安全」と見なされるローカル開発環境のホスト名を定義します。
// mapを使用することで、ルックアップが高速になり、将来的に新しいホスト名の追加が容易になります。
var localdevHostnames = map[string]struct{}{
	"localhost": {},
	"127.0.0.1": {},
	"::1":       {},
	// 将来的に "host.docker.internal" などをここに追加可能
}

// IsSecureServiceURL は、提供されたServiceURLがHTTPSスキームを使用しているか、
// またはローカル開発環境の例外に該当するかを判断します。
// これは、WebアプリケーションでのクッキーのSecure属性設定などのセキュリティチェックに使用されます。
func IsSecureServiceURL(serviceURL string) bool {
	u, err := url.Parse(serviceURL)
	if err != nil {
		// パースできないURLは安全ではないと判断
		return false
	}

	if u.Scheme == "https" {
		return true
	}

	if u.Scheme == "http" {
		// ホスト名を小文字に変換し、パッケージレベルの許可リストでチェック
		hostname := strings.ToLower(u.Hostname())

		_, isLocaldev := localdevHostnames[hostname]
		return isLocaldev
	}

	// その他のスキーム (ftp, file, etc.) は安全ではないと判断
	return false
}
