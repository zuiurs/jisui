package comic

import "testing"

func TestGetDestImagePath(t *testing.T) {
	tests := []struct {
		srcPath  string
		output   string
		ext      string
		isDir    bool
		expected string
	}{
		{
			"/path/to/file",
			"hello.jpg",
			"png",
			false,
			"hello.jpg",
		},
		{
			"/path/to/file",
			"",
			"png",
			false,
			"/path/to/file.png",
		},
		{
			"/path/to/file.jpg",
			"",
			"png",
			false,
			"/path/to/file.png",
		},
		{
			"/path/to/dir/file",
			"hello",
			"png",
			true,
			"hello/file.png",
		},
		{
			"/path/to/dir/file",
			"",
			"png",
			true,
			"/path/to/dir/file.png",
		},
		{
			"/path/to/dir/file.jpg",
			"",
			"png",
			true,
			"/path/to/dir/file.png",
		},
		{
			"/path/to/dir",
			"hello.pdf",
			"pdf",
			false,
			"hello.pdf",
		},
		{
			"/path/to/dir",
			"",
			"pdf",
			false,
			"/path/to/dir.pdf",
		},
	}

	for i, tt := range tests {
		if path := getDestImagePath(tt.srcPath, tt.output, tt.ext, tt.isDir); path != tt.expected {
			t.Fatalf("%d: expected=%s, got=%s", i, tt.expected, path)
		}
	}
}
