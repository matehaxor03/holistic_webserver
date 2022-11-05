package main

import (
	"log"
	"net/http"
)

func ProcessRequest(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world"))
}

func main() {
	buildHandler := http.FileServer(http.Dir("static"))
	http.Handle("/", buildHandler)

	http.HandleFunc("/api", ProcessRequest)

	err := http.ListenAndServeTLS(":5001", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
