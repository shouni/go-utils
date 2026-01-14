package urlpath

import (
	"testing"
)

func TestIsRemoteURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want bool
	}{
		{"GCS URI", "gs://my-bucket/path", true},
		{"S3 URI", "s3://my-bucket/path", true},
		{"Local path", "/usr/local/bin", false},
		{"Relative path", "./local/file", false},
		{"HTTP URL", "https://example.com", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRemoteURI(tt.uri); got != tt.want {
				t.Errorf("IsRemoteURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveOutputPath(t *testing.T) {
	tests := []struct {
		name     string
		baseDir  string
		fileName string
		want     string
		wantErr  bool
	}{
		{"GCS successful join", "gs://bucket/dir", "image.png", "gs://bucket/dir/image.png", false},
		{"S3 successful join", "s3://bucket/dir", "data.json", "s3://bucket/dir/data.json", false},
		{"Local successful join", "/tmp", "test.txt", "/tmp/test.txt", false},
		{"GCS with trailing slash", "gs://bucket/dir/", "image.png", "gs://bucket/dir/image.png", false},
		{"Invalid GCS URI", "gs://%%invalid", "file.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveOutputPath(tt.baseDir, tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveOutputPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		rawPath string
		want    string
	}{
		{"GCS file", "gs://bucket/images/char.png", "gs://bucket/images/"},
		{"HTTPS URL", "https://example.com/assets/logo.svg", "https://example.com/assets/"},
		{"Local absolute path", "/home/user/data.txt", "/home/user/"},
		{"Local relative path", "dir/file.png", "dir/"},
		{"Current directory", "file.txt", "./"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveBaseURL(tt.rawPath); got != tt.want {
				t.Errorf("ResolveBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateIndexedPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		index    int
		want     string
		wantErr  bool
	}{
		{"Normal png", "path/to/image.png", 1, "path/to/image_1.png", false},
		{"Large index", "image.jpg", 99, "image_99.jpg", false},
		{"No extension", "README", 2, "README_2", false},
		{"GCS path", "gs://bucket/asset.png", 5, "gs://bucket/asset_5.png", false},
		{"Zero index error", "image.png", 0, "", true},
		{"Negative index error", "image.png", -1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateIndexedPath(tt.basePath, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateIndexedPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateIndexedPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
