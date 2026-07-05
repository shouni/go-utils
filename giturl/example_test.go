package giturl_test

import (
	"fmt"

	"github.com/shouni/go-utils/giturl"
)

func ExampleGetRepositoryPath() {
	fmt.Println(giturl.GetRepositoryPath("git@github.com:owner/repo.git"))
	// Output: owner/repo
}

func ExampleGenerateGCSKeyName() {
	fmt.Println(giturl.GenerateGCSKeyName("git@github.com:owner/repo.git"))
	// Output: github-com-owner-repo-b1cd17c6
}
