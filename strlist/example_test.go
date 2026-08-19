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
