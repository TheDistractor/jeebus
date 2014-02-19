package jeebus

import (
	"net/http"
)

func init() {
	ad := http.FileServer(http.Dir(Settings.CommonDir))
	http.Handle("/common/", http.StripPrefix("/common/", ad))
}
