package main

import (
	"bytes"
	libsass "github.com/wellington/go-libsass"
	"net/http"
	"os"
)

func compileSass() *bytes.Reader {
	var buffer bytes.Buffer

	file, err := os.Open("template/css/style.scss")
	if err != nil {
		bail(err)
	}

	compiler, err := libsass.New(&buffer, file)
	if err != nil {
		bail(err)
	}

	includePaths := []string{"template/css/"}
	err = compiler.Option(libsass.IncludePaths(includePaths))
	if err != nil {
		bail(err)
	}

	if err := compiler.Run(); err != nil {
		bail(err)
	}

	return bytes.NewReader(buffer.Bytes())
}

func stylesheet(buffer *bytes.Reader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		buffer.WriteTo(w)
		buffer.Seek(0, 0)
	}
}
