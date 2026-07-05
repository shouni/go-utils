package iohandler_test

import (
	"fmt"

	"github.com/shouni/go-utils/iohandler"
)

func ExampleWriteOutputString() {
	// filename が空の場合は標準出力に書き込まれます
	if err := iohandler.WriteOutputString("", "hello, iohandler"); err != nil {
		fmt.Println("error:", err)
	}
	// Output: hello, iohandler
}
