package envutil_test

import (
	"fmt"

	"github.com/shouni/go-utils/envutil"
)

func ExampleGetEnv() {
	value := envutil.GetEnv("MISSING_ENV_KEY", "default-value")
	fmt.Println(value)
	// Output: default-value
}

func ExampleGetEnvAsBool() {
	value := envutil.GetEnvAsBool("MISSING_ENV_KEY", true)
	fmt.Println(value)
	// Output: true
}

func ExampleGetEnvAsInt() {
	value := envutil.GetEnvAsInt("MISSING_ENV_KEY", 42)
	fmt.Println(value)
	// Output: 42
}
