// Package strlist は、設定値として読み込んだ文字列リストを整えます。
//
// 対象は「分割済みのリスト」で、カンマ区切り文字列の分割そのものは扱いません。
// それは設定ライブラリ（caarlos0/env など）の担当だからです。
package strlist

import "strings"

// Normalize は、設定値として読み込んだ文字列リストの表記ゆれを整えます。
// 各要素の前後の空白を落とし、空要素と重複を捨て、最初に現れた順序は保ちます。
//
// 分割器は `strings.Split` するだけで空白も空要素も落とさないため、`"a, b,,a"` は
// `["a", " b", "", "a"]` のまま渡ってきます。これが下流へ流れると、許可リストの照合が
// 空白付きの要素で外れたり、選択肢に同じ項目が二度並んだりします。
//
// 既定値で埋めることはせず、入力が空なら空のスライスを返します。空のまま通してよいかは
// 利用側の検証が決めるべき事柄で、ここで既定値を差し込むと設定漏れが隠れるためです。
func Normalize(values []string) []string {
	return normalize(values, nil)
}

// NormalizeFold は、Normalize に加えて各要素を小文字化し、大文字小文字の違いを
// 重複とみなします。`["Example.com", "EXAMPLE.com"]` は `["example.com"]` になります。
//
// ホスト名やメールドメイン、識別子の許可リストのように、値そのものが大文字小文字を
// 区別しない設定で使います。照合側も同じく小文字化してから比較してください。
// 大文字小文字に意味がある値（トークンや API キーなど）には Normalize を使います。
func NormalizeFold(values []string) []string {
	return normalize(values, strings.ToLower)
}

// normalize は Normalize と NormalizeFold の共通処理です。
// fold が nil でなければ、空白を落とした後の各要素へ適用します。
func normalize(values []string, fold func(string) string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if fold != nil {
			v = fold(v)
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		normalized = append(normalized, v)
	}

	return normalized
}
