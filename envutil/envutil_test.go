package envutil

import (
	"testing"
)

func TestGetEnv(t *testing.T) {
	const key = "TEST_STRING_KEY"
	const defaultValue = "default_val"

	tests := []struct {
		name     string
		envValue string
		exists   bool
		want     string
	}{
		{"Environment variable exists", "real_value", true, "real_value"},
		{"Environment variable is empty", "", true, ""},
		{"Environment variable does not exist", "", false, "default_val"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exists {
				t.Setenv(key, tt.envValue)
			} else {
				// 環境変数が確実に存在しない状態にする
				t.Setenv(key, "") // 一旦セット
				// os.Unsetenv は t.Setenv の自動クリーンアップを妨げる可能性があるため
				// 基本的には t.Setenv を使わないケース（defaultValue）として扱う
			}

			// テスト対象の実行
			// 存在しないケースは key を変えて確実に defaultValue を狙う
			testKey := key
			if !tt.exists {
				testKey = "NON_EXISTENT_KEY_XYZ"
			}

			if got := GetEnv(testKey, defaultValue); got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	const key = "TEST_BOOL_KEY"

	tests := []struct {
		name         string
		envValue     string
		exists       bool
		defaultValue bool
		want         bool
	}{
		{"Exists and true", "true", true, false, true},
		{"Exists and 1 as true", "1", true, false, true},
		{"Exists and false", "false", true, true, false},
		{"Exists and 0 as false", "0", true, true, false},
		{"Invalid value returns default (true)", "not_a_bool", true, true, true},
		{"Invalid value returns default (false)", "not_a_bool", true, false, false},
		{"Not exists returns default (true)", "", false, true, true},
		{"Not exists returns default (false)", "", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testKey := key
			if tt.exists {
				t.Setenv(key, tt.envValue)
			} else {
				testKey = "NON_EXISTENT_BOOL_KEY"
			}

			if got := GetEnvAsBool(testKey, tt.defaultValue); got != tt.want {
				t.Errorf("GetEnvAsBool() = %v, want %v (env=%q, exists=%v)", got, tt.want, tt.envValue, tt.exists)
			}
		})
	}
}
