package server

import (
	"fmt"
	"log"
	"mime"
	"net/http"
)

func GetMux() *http.ServeMux {
	mux := http.NewServeMux()
	initMux(mux)
	return mux
}

func ping(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprint(w, "Pong!\n")
	if err != nil {
		log.Fatal("Error writing response")
	}
}

func determineFT(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mimeType, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		http.Error(w, "Incorrect content type, ignoring request", 400)
		return
	}

	if mimeType != "multipart/form-data" {
		http.Error(w, "Incorrect content type, ignoring request", 400)
		fmt.Printf("bad `determine ft` request of type: %s\n", mimeType)
		return
	}
	r, err := req.MultipartReader()
	if err != nil {
		log.Fatal("Error parsing multipart form")
	}

	buf := make([]byte, 512)
	part, err := r.NextRawPart()
	if err != nil {
		log.Fatal("Error occured while reading multipart form data")
	}
	defer part.Close()
	n, err := part.Read(buf)
	if err != nil {
		log.Fatal("Error reading part data")
	}

	fmt.Fprintf(w, "Detected filetype: %s\n", http.DetectContentType(buf[:n]))
}

func initMux(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", ping)
	mux.HandleFunc("POST /ft", determineFT)
	mux.HandleFunc("POST /upload", upload)
	mux.HandleFunc("GET /f/{filePath...}", download)
}
