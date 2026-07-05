package urlpath_test

import (
	"fmt"

	"github.com/shouni/go-utils/urlpath"
)

func ExampleResolvePath() {
	path, err := urlpath.ResolvePath("gs://my-bucket/images", "photo.png")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(path)
	// Output: gs://my-bucket/images/photo.png
}

func ExampleIsRemoteURI() {
	fmt.Println(urlpath.IsRemoteURI("gs://my-bucket/photo.png"))
	fmt.Println(urlpath.IsRemoteURI("/local/photo.png"))
	// Output:
	// true
	// false
}

func ExampleGenerateIndexedPath() {
	path, err := urlpath.GenerateIndexedPath("gs://my-bucket/images/photo.png", 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(path)
	// Output: gs://my-bucket/images/photo_2.png
}
