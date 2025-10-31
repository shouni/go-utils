package text_test

import (
	"github.com/shouni/go-utils/text" // パッケージ名を 'text' と仮定
	"testing"
)

func TestCleanStringFromEmojis(t *testing.T) { // 関数名を CleanStringFromEmojis に合わせることを推奨
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "絵文字なし",
			input:    "これは通常のテキストです。",
			expected: "これは通常のテキストです。",
		},
		{
			name:     "標準的な絵文字を含む",
			input:    "こんにちは😃世界🌏",
			expected: "こんにちは世界",
		},
		{
			name:     "肌の色の修飾子付き絵文字を含む",
			input:    "👍🏻 いいね！",
			expected: "いいね！", // 修正: 先頭のスペースが削除されるため
		},
		{
			name:     "旗の絵文字を含む",
			input:    "日本の旗🇯🇵とアメリカの旗🇺🇸",
			expected: "日本の旗とアメリカの旗",
		},
		{
			name:     "結合された絵文字（ZWGシーケンス）を含む",
			input:    "👨‍👩‍👧‍👦 家族の絵文字",
			expected: "家族の絵文字", // 修正: 先頭のスペースが削除されるため
		},
		{
			name:     "数字と句読点のみ",
			input:    "123,456.78",
			expected: "123,456.78",
		},
		{
			name:     "絵文字と空白文字のみ",
			input:    " 🎉  ✨ ",
			expected: "", // 修正: 結果が空白文字のみになるため
		},
		{
			name:     "空文字列",
			input:    "",
			expected: "",
		},
		{
			name:     "絵文字以外の特殊記号",
			input:    "¥$€£&@%",
			expected: "¥$€£&@%",
		},
		{
			name:     "文頭文末と連続する空白を含むテキスト",
			input:    "  テスト  テキスト   です。 ",
			expected: "テスト テキスト です。", // 空白整理の動作を確認
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ここでテスト対象の関数名を CleanStringFromEmojis に修正
			actual := text.CleanStringFromEmojis(tt.input)
			if actual != tt.expected {
				t.Errorf("CleanStringFromEmojis(%q) = %q, 期待値 %q", tt.input, actual, tt.expected)
			}
		})
	}
}
