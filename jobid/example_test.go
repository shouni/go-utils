package jobid_test

import (
	"fmt"

	"github.com/shouni/go-utils/jobid"
)

func ExampleValidate() {
	fmt.Println(jobid.Validate("20260725123456-abcd1234"))
	fmt.Println(jobid.Validate("../etc/passwd"))
	// Output:
	// <nil>
	// invalid job id: "../etc/passwd"
}

func ExampleSanitize() {
	// 外部入力に紛れ込んだパス要素を落としてから検証します。
	id, err := jobid.Sanitize("../../20260725123456-abcd1234")
	fmt.Println(id, err)
	// Output: 20260725123456-abcd1234 <nil>
}

func ExampleIsValid() {
	fmt.Println(jobid.IsValid("video-recipe-20260725-150405-a1b2c3d4"))
	fmt.Println(jobid.IsValid("-leading-hyphen"))
	// Output:
	// true
	// false
}
