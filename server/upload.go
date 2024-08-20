package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"file-cellar/config"
	"file-cellar/db"
	"file-cellar/storage"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func computeHash(hasher hash.Hash, data io.Reader) (string, error) {
	if _, err := io.Copy(hasher, data); err != nil {
		return "", err
	}

	hashedBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashedBytes), nil
}

func detectFileType(f io.ReadSeeker) string {
	buf := make([]byte, 512)
	t := http.DetectContentType(buf)
	f.Seek(0, io.SeekStart)

	return t
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

	// TODO: use hash strategy set in server config
	hash, err := computeHash(md5.New(), formFile)
	if err != nil || hash == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to compute file hash: %v :%s", err, r.RemoteAddr)
		return
	}
	if _, err = formFile.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to reset file after computing its hash: %v : %s", err, r.RemoteAddr)
	}

	fileType := detectFileType(formFile)

	relPath, err := storage.GetRelPath(fileFormHeader.Filename, hash, uploadTime)
	if err != nil {
		http.Error(w, "Missing Filename in upload", http.StatusBadRequest)
		log.Println("Missing filename for upload: ", r.RemoteAddr)
		return
	}

	fInfo := storage.FileInfo{
		Name:            fileFormHeader.Filename,
		Hash:            hash,
		Type:            fileType,
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
