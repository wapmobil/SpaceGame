package web

import (
	"io/fs"
	"os"
	"path/filepath"
)

var subFS fs.FS

func init() {
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	webDir := filepath.Join(exeDir, "web")
	subFS = os.DirFS(webDir)
}

func Sub() fs.FS {
	return subFS
}
