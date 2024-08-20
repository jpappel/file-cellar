package server

import (
	"file-cellar/config"
	"file-cellar/db"
	"file-cellar/storage"
	"log"
	"net/http"
)

func download(w http.ResponseWriter, r *http.Request) {
	// TODO: get relpath
	path := r.PathValue("filePath")
	if path == "" {
		http.Error(w, "Missing path", http.StatusBadRequest)
		log.Println("Missing file path: ", r.RemoteAddr)
        return
	}

	manager, err := db.GetManager(config.Server["DBURL"], config.DB_PRAGMAS)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting manager for: %s\n", r.RemoteAddr)
		return
	}
	defer manager.Close()

	ctx := r.Context()

	fInfo, err := manager.GetFile(ctx, path)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting file info: %v : %s\n", err, r.RemoteAddr)
		return
	}

	f, redirectUrl, err := fInfo.Bin.Get(ctx, storage.FileIdentifier(fInfo.RelPath))
	if err != nil {
		// TODO: handle errors from Bin Get
	}
	defer f.Close()

	if fInfo.Bin.Redirect {
		http.Redirect(w, r, redirectUrl, http.StatusPermanentRedirect)
		return
	}

	http.ServeContent(w, r, fInfo.Name, fInfo.UploadTimestamp, f)
}
