package ui

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type UI struct {
	BaseDir string
}

func (ui *UI) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	assetPath := os.Getenv("ASSET_PATH")

	if len(assetPath) == 0 {
		assetPath = "assets/default"
	}

	ver := struct {
		Version string
	}{
		Version: assetPath,
	}

	tmpPath := filepath.Join(ui.BaseDir, "./ui/tmpl/index.html")

	t := template.New("some template")
	t, err := t.ParseFiles(tmpPath)

	if err != nil {
		log.Println(tmpPath)
	}

	t.ExecuteTemplate(res, "index", ver)
}
