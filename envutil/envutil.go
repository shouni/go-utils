package envutil

import (
	"os"
	"strings"
)

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返します。
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsBool は環境変数からbool値を読み込みます。
// 環境変数が設定されていないか、"true" 以外の値の場合はdefaultValueを返します。
func getEnvAsBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return strings.ToLower(val) == "true"
}
