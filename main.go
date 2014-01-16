package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("public")))

	println("listening on port 3333")
	log.Fatal(http.ListenAndServe("localhost:3333", mux))
}
