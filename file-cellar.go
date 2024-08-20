package main

import (
	"context"
	"file-cellar/config"
	"file-cellar/db"
	"file-cellar/server"
	"file-cellar/storage"
	"fmt"
	"log"
	"net/http"
)

func main() {

	ctx := context.Background()
	manager, err := db.GetManager(config.Server["DBURL"], config.DB_PRAGMAS)
	if err != nil {
		log.Panicf("Unable to get manager: %v\n", err)
	}
	defer manager.Close()

	err = manager.Init()
	if err != nil {
		log.Panicf("Failed to initialize tables: %v\n", err)
	}

	localDriver := storage.NewLocalDriver()
	if !manager.AddDriver(ctx, localDriver) {
		log.Panicf("Failed to register local driver: %v\n", err)
	}

	bin := &storage.Bin{
		Name:     "testing bin",
		Driver:   localDriver,
		Redirect: false,
	}
	bin.Path.External = "foobar"
	bin.Path.Internal = "./testing/files"

	_, err = manager.AddBin(ctx, bin, bin.Driver.Id())
	if err != nil {
		log.Panicf("Error adding bin: %v\n", err)
	}

	const PORT uint = 8080
	mux := server.GetMux()
	log.Printf("Listening on %d\n", PORT)
	http.ListenAndServe(":"+fmt.Sprint(PORT), mux)
}
