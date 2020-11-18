package main

import (
	"io"
	"net/http"
)

// WebStartUp implements a dumb web server
func WebStartUp() {
	http.HandleFunc("/", servePage)
	http.ListenAndServe(":8080", nil)
}

func servePage(writer http.ResponseWriter, reqest *http.Request) {
	io.WriteString(writer, "Hello world!")
}
