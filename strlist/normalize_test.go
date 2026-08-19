package strlist_test

import (
	"reflect"
	"testing"

	"github.com/shouni/go-utils/strlist"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "整った入力はそのまま",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			// 設定ライブラリの分割は前後の空白を落としません。
			name:     "要素の前後の空白を落とす",
			input:    []string{"  apple  ", "banana", "  cherry "},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			// 値の末尾や連続したカンマから生まれます。
			name:     "空要素を捨てる",
			input:    []string{"", "apple", "", "banana", ""},
			expected: []string{"apple", "banana"},
		},
		{
			name:     "空白のみの要素も捨てる",
			input:    []string{" ", "apple", "   "},
			expected: []string{"apple"},
		},
		{
			// トリム後に同じになる要素も重複として扱います。
			name:     "重複を捨て、最初に現れた順序を保つ",
			input:    []string{"banana", "apple", " banana ", "cherry", "apple"},
			expected: []string{"banana", "apple", "cherry"},
		},
		{
			name:     "空スライスは空スライス",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "nil は空スライス",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "すべて空または空白なら空スライス",
			input:    []string{" ", "", "  "},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := strlist.Normalize(tt.input)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Normalize(%q) = %v, 期待値 %v", tt.input, actual, tt.expected)
			}
		})
	}
}

// 入力スライスを書き換えないこと。設定の生値をログに出す用途などで、
// 呼び出し側が元の値を保持している場合があります。
func TestNormalizeDoesNotMutateInput(t *testing.T) {
	input := []string{" a ", "", "a", "b"}
	want := []string{" a ", "", "a", "b"}

	strlist.Normalize(input)

	if !reflect.DeepEqual(input, want) {
		t.Errorf("入力が書き換えられました: %v, want %v", input, want)
	}
}
