package jeebus

import (
	"net/http"
	"os"
)

func init() {
	err := os.MkdirAll(Settings.FilesDir, 0755)
	Check(err)

	fs := http.FileServer(http.Dir(Settings.FilesDir))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
}
