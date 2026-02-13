package static

import (
	"embed"
	"io/fs"
)

//go:embed css js
var files embed.FS

// FS returns the embedded static filesystem.
func FS() fs.FS {
	return files
}
