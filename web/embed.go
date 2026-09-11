package web

import (
	"embed"
	"io/fs"
)

//go:embed build/*
var buildFS embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(buildFS, "build")
}
