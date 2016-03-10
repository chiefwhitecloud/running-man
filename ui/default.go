package ui

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type UI struct {
	Template []byte
	BaseDir  string
}

func (ui *UI) GetDefaultTemplate(res http.ResponseWriter, req *http.Request) {

	ver := struct {
		Version string
	}{
		Version: "default",
	}

	tmpPath := filepath.Join(ui.BaseDir, "./ui/tmpl/index.html")

	log.Println(tmpPath)

	t := template.New("some template") // Create a template.
	//t, _ = t.Parse("hello {{.Version}}!")
	t, err := t.ParseFiles(tmpPath) // Parse template file

	if err != nil {
		log.Println(tmpPath)
	}

	t.ExecuteTemplate(res, "index", ver)

	//t.Execute(res, ver) // merge.

}
