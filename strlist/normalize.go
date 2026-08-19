// Package strlist は、設定値として読み込んだ文字列リストを整えます。
//
// 対象は「分割済みのリスト」で、カンマ区切り文字列の分割そのものは扱いません。
// それは設定ライブラリ（caarlos0/env など）の担当だからです。
package strlist

import "strings"

// Normalize は、設定値として読み込んだ文字列リストの表記ゆれを整えます。
// 各要素の前後の空白を落とし、空要素と重複を捨て、最初に現れた順序は保ちます。
//
// カンマ区切りの分割そのものは呼び出し側（設定ライブラリなど）の担当です。
// 分割器はたいてい `strings.Split` するだけで前後の空白も空要素も落とさないため、
// その後始末をここが引き受けます。`"a, b,,a"` のような値が
// `["a", " b", "", "a"]` のまま下流へ流れると、許可リストの照合が
// 空白付きの要素で外れたり、選択肢に同じ項目が二度並んだりします。
//
// 既定値で埋めることはせず、入力が空なら空のスライスを返します。空のまま通してよいかは
// 利用側の検証が決めるべき事柄で、ここで既定値を差し込むと設定漏れが隠れるためです。
func Normalize(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		normalized = append(normalized, v)
	}

	return normalized
}
