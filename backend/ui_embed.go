package backend

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed ui/*
var UI embed.FS

// LoadUIFileSystem gets the sub-directory containing the actual index.html
func LoadUIFileSystem() http.FileSystem {
	subFS, err := fs.Sub(UI, "ui")
	if err != nil {
		log.Fatal("Falha ao embutir filesystem do UI: ", err)
	}
	return http.FS(subFS)
}
