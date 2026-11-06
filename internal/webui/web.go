package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embedded embed.FS

func Files() fs.FS {
	f, e := fs.Sub(embedded, "dist")
	if e != nil {
		panic(e)
	}
	return f
}
