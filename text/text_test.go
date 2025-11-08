package text_test

import (
	"testing"

	"github.com/shouni/go-utils/text"
)

func TestCleanStringFromEmojis(t *testing.T) {
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
			actual := text.CleanStringFromEmojis(tt.input)
			if actual != tt.expected {
				t.Errorf("CleanStringFromEmojis(%q) = %q, 期待値 %q", tt.input, actual, tt.expected)
			}
		})
	}
}

// TestTruncate 関数のテスト
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		suffix   string
		expected string
	}{
		{
			name:     "クリティカル: maxLenが負の値 (-1)",
			input:    "テストテキスト",
			maxLen:   -1,
			suffix:   "...",
			expected: "...", // 負の値は0として扱われ、サフィックスのみが返される
		},
		{
			name:     "クリティカル: maxLenがゼロ (0)",
			input:    "テストテキスト",
			maxLen:   0,
			suffix:   "...",
			expected: "...", // 0として扱われ、サフィックスのみが返される
		},
		{
			name:     "クリティカル: maxLenがゼロ (0) かつ空文字列",
			input:    "",
			maxLen:   0,
			suffix:   "...",
			expected: "", // 入力が空の場合は、サフィックスも付加されない
		},
		{
			name:     "最大長より短い文字列",
			input:    "Hello",
			maxLen:   10,
			suffix:   "...",
			expected: "Hello",
		},
		{
			name:     "最大長と等しい文字列",
			input:    "HelloWorld",
			maxLen:   10,
			suffix:   "...",
			expected: "HelloWorld",
		},
		{
			name:     "最大長を超える文字列 (サフィックスあり)",
			input:    "This is a long text.",
			maxLen:   10,
			suffix:   "...",
			expected: "This is a...",
		},
		{
			name:     "最大長を超える文字列 (サフィックスなし)",
			input:    "This is a long text.",
			maxLen:   10,
			suffix:   "",
			expected: "This is a",
		},
		{
			name:     "切り詰めた末尾がスペースの場合",
			input:    "ABCDEFGHI JKLM",
			maxLen:   11,
			suffix:   "...",
			expected: "ABCDEFGHI J...",
		},
		{
			name:     "空文字列",
			input:    "",
			maxLen:   5,
			suffix:   "...",
			expected: "",
		},
		{
			name:     "マルチバイト文字を含む (rune長で切り詰め)",
			input:    "あいうえお",
			maxLen:   3,
			suffix:   "...",
			expected: "あいう...",
		},
		{
			name:     "マルチバイト文字を最大長より多く指定",
			input:    "あいうえお",
			maxLen:   7,
			suffix:   "...",
			expected: "あいうえお",
		},
		{
			name:     "末尾に空白がある日本語",
			input:    "テストテキスト　　です。 ",
			maxLen:   6,
			suffix:   "...",
			expected: "テストテキス...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := text.Truncate(tt.input, tt.maxLen, tt.suffix)
			if actual != tt.expected {
				t.Errorf("Truncate(%q, %d, %q) = %q, 期待値 %q", tt.input, tt.maxLen, tt.suffix, actual, tt.expected)
			}
		})
	}
}
