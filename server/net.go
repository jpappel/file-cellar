package server

import (
	"context"
	"file-cellar/config"
	"file-cellar/db"
	"file-cellar/storage"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"
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

func _upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10e6)
	if err != nil {
		http.Error(w, "Error Parsing Form", 400)
		log.Printf("Error Occured: %v", err)
		return
	}

	file := r.MultipartForm.File["file"][0]
	o, err := file.Open()
	defer o.Close()

	dest, err := os.Create(file.Filename)
	if err != nil {
		http.Error(w, "Error Creating file", 500)
		log.Printf("Error creating file: %s", file.Filename)
		return
	}
	defer dest.Close()

	io.Copy(dest, o)

	fmt.Fprintln(w, "File uploaded succesfully!")
	log.Printf("File %s recieved from %s\n", file.Filename, r.RemoteAddr)
}

func initMux(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", ping)
	mux.HandleFunc("POST /ft", determineFT)
	mux.HandleFunc("POST /upload", upload)
}

func upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10e6)
	if err != nil {
		// TODO: use correct http status code
		http.Error(w, "Error Parsing MultiPartForm data", 400)
		log.Printf("Parsing Multipart form failed: %s\n", r.RemoteAddr)
		return
	}

	formFile, fileFormHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error finding file in upload, did you include it under the name `file`?", http.StatusBadRequest)
		log.Printf("Missing file in upload: %s\n", r.RemoteAddr)
		return
	}

	manager, err := db.GetManager(config.Server["DBURL"], config.DB_PRAGMAS)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting manager for: %s\n", r.RemoteAddr)
		return
	}
	defer manager.Close()

	binId, err := strconv.ParseInt(r.FormValue("binId"), 10, 64)
	if err != nil || binId < 0 {
		http.Error(w, fmt.Sprintf("Bad binId `%s`, it should be a positive integer", r.FormValue("binId")), http.StatusBadRequest)
		log.Printf("Bad bin id `%s`: %s", r.FormValue("binId"), r.RemoteAddr)
		return
	}

	ctx := r.Context()
	bin, err := manager.GetBin(ctx, binId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not find bin with id `%d`", binId), http.StatusBadRequest)
		log.Printf("No bin with id `%d`: %s\n", binId, r.RemoteAddr)
		return
	}

	uploadTime := time.Now()
	hash := "1234" // TODO: compute hash
	relPath, err := storage.GetRelPath(fileFormHeader.Filename, hash, uploadTime)

	fInfo := storage.FileInfo{
		Name:            fileFormHeader.Filename,
		Hash:            hash,
		Size:            fileFormHeader.Size,
		RelPath:         relPath,
		UploadTimestamp: uploadTime,
		Bin:             bin,
	}
	f := &storage.File{
		Data:     formFile,
		FileInfo: fInfo,
	}

	if err = manager.AddFile(ctx, &fInfo); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error adding file info to database: %s\n", r.RemoteAddr)
		return
	}

	if err = f.Bin.Upload(ctx, f); err != nil {
		http.Error(w, "Error while saving file", http.StatusInternalServerError)
		log.Println("Saving uploaded file failed: ", err)
		log.Println("Attempting cleanup: ", r.RemoteAddr)
		err = f.Bin.Delete(context.TODO(), storage.FileIdentifier(f.RelPath))
		if err != nil {
			log.Panicf("Failed removing file info from database :%v\n%v\n", err, fInfo)
		}
		return
	}

	// FIXME: use correct accesible url
	fmt.Fprintf(w, "%s\n", relPath)
	log.Printf("File uploaded %s from %s", relPath, r.RemoteAddr)
}
