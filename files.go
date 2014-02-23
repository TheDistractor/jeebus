package jeebus

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

func init() {
	err := os.MkdirAll(Settings.FilesDir, 0755)
	Check(err)

	fs := http.FileServer(http.Dir(Settings.FilesDir))
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	Define("fetch", func(orig string, args []interface{}) interface{} {
		return Fetch(args[0].(string))
	})

	Define("store", func(orig string, args []interface{}) interface{} {
		var data []byte
		if len(args) > 1 {
			data = []byte(args[1].(string))
		}
		return Store(args[0].(string), data)
	})

	Define("file-list", func(orig string, args []interface{}) interface{} {
		return FileList(args[0].(string), args[1].(bool))
	})
}

func PathIsSafe(s string) bool {
	// TODO: still a bit ad-hoc, with probably lots of holes in these checks
	s = path.Clean(s)
	return s != "" && !strings.HasPrefix(s, ".") && !strings.HasSuffix(s, "/")
}

func Fetch(filename string) (data []byte) {
	if PathIsSafe(filename) {
		data, _ = ioutil.ReadFile(Settings.FilesDir + "/" + filename)
	}
	return
}

func Store(filename string, body []byte) (err error) {
	if PathIsSafe(filename) {
		fpath := Settings.FilesDir + "/" + filename
		if len(body) > 0 {
			if err = os.MkdirAll(path.Dir(fpath), 0755); err == nil {
				err = ioutil.WriteFile(fpath, body, 0666)
			}
		} else {
			// iterate to automatically clean out empty parent dirs
			for {
				err = os.Remove(fpath)
				if err != nil {
					break
				}
				filename = path.Dir(filename)
				if filename == "." {
					break
				}
				fpath = Settings.FilesDir + "/" + filename
				all, _ := ioutil.ReadDir(fpath)
				if len(all) > 0 {
					break
				}
			}
		}
	}
	return
}

func FileList(dirname string, dir bool) (files []string) {
	if dirname == "." || PathIsSafe(dirname) {
		all, _ := ioutil.ReadDir(Settings.FilesDir + "/" + dirname)
		for _, f := range all {
			if f.IsDir() == dir {
				files = append(files, f.Name())
			}
		}
	}
	return
}
