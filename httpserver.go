package jeebus

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

func init() {
	err := os.MkdirAll(Settings.AppDir, 0755)
	Check(err)

	fs := http.FileServer(http.Dir(Settings.AppDir))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if path.Ext(r.URL.Path) == "" {
			r.URL.Path = "/" // only serve name.ext as real files for SPA's
		}
		fs.ServeHTTP(w, r)
	})

	ba := http.FileServer(http.Dir(Settings.BaseDir))
	http.Handle("/base/", http.StripPrefix("/base/", ba))
}

func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.DefaultServeMux.ServeHTTP(rw, req)
}

func RunHttpServer() {
	port := fmt.Sprintf(":%d", Settings.Port)

	if Settings.CertFile != "" && Settings.KeyFile != "" {
		log.Print("starting TLS server on https://localhost" + port)
		log.Fatal(http.ListenAndServeTLS(port,
			Settings.CertFile, Settings.KeyFile, nil))
	} else {
		log.Print("starting HTTP server on http://localhost" + port)
		log.Fatal(http.ListenAndServe(port, nil))
	}
}
