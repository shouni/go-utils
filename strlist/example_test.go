package strlist_test

import (
	"fmt"

	"github.com/shouni/go-utils/strlist"
)

func ExampleNormalize() {
	// 設定ライブラリがカンマで分割しただけの値を渡します。
	items := strlist.Normalize([]string{"go", " rust ", "", "go", "python"})
	fmt.Println(items)
	// Output: [go rust python]
}

func ExampleNormalizeFold() {
	// ホスト名のように大文字小文字を区別しない設定値を揃えます。
	hosts := strlist.NormalizeFold([]string{"Example.com", " EXAMPLE.com ", "api.example.com"})
	fmt.Println(hosts)
	// Output: [example.com api.example.com]
}
