// Package iohandler provides utility functions for reading content from files or stdin
// and writing content to files or stdout.
package iohandler

import (
	"fmt"
	"io"
	"os"
)

// ReadInput reads content from a file or stdin.
// NOTE: Large file support is currently deferred. The entire file is read into memory.
func ReadInput(filename string) ([]byte, error) {
	if filename != "" {
		// Log removed: Library functions should not directly output progress messages to os.Stderr.
		return os.ReadFile(filename)
	}

	// Log removed: Library functions should not directly output progress messages to os.Stderr.
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		// Error message translated to English.
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}
	return content, nil
}

// WriteOutput writes content to a file or stdout.
// NOTE: Large file support is currently deferred. The entire content is written from memory.
func WriteOutput(filename string, content []byte) error { // content string から content []byte に変更
	if filename != "" {
		// Log removed: Library functions should not directly output completion messages to os.Stderr.
		return os.WriteFile(filename, content, 0644) // []byte(content) の変換が不要に
	}

	// Log removed: Library functions should not directly output result separators to os.Stderr.

	// os.Stdout.Write を使用し、fmt.Fprint や fmt.Fprintln のオーバーヘッドを回避
	_, err := os.Stdout.Write(content)
	if err != nil {
		// Error message translated to English.
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	return nil
}
