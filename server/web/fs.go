package web

import (
	"embed"
	"io/fs"
)

//go:embed *
var files embed.FS

// Sub returns a sub-FS rooted at the embedded web directory.
func Sub() fs.FS {
	sub, err := fs.Sub(files, ".")
	if err != nil {
		panic(err)
	}
	return sub
}
