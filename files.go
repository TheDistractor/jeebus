package jeebus

import (
	"io/ioutil"
	"net/http"
	"os"
)

func init() {
	err := os.MkdirAll(Settings.FilesDir, 0755)
	Check(err)

	fs := http.FileServer(http.Dir(Settings.FilesDir))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
}

func Fetch(filename string) []byte {
	// TODO: this isn't safe if the filename uses a nasty path!
	data, _ := ioutil.ReadFile(Settings.FilesDir + "/" + filename)
	return data
}

func Store(filename string, body []byte) error {
	// TODO: this isn't safe if the filename uses a nasty path!
	fpath := Settings.FilesDir + "/" + filename
	if len(body) > 0 {
		return ioutil.WriteFile(fpath, body, 0666)
	} else {
		return os.Remove(fpath)
	}
}
