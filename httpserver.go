package jeebus

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func init() {
	err := os.MkdirAll(Settings.AppDir, 0755)
	Check(err)

	http.Handle("/", http.FileServer(http.Dir(Settings.AppDir)))
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
