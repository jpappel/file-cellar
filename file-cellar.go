package main

import (
	"file-cellar/db"
	"log"
)

func main() {
	const connStr = "./testing.db"
	pool, err := db.GetPool(connStr, db.SQLITE_DEFAULT_PRAGMAS)
	if err != nil {
		log.Fatal("Failed to get database connection", err)
	}
	defer db.ClosePool(connStr)

	err = db.InitTables(pool)
	if err != nil {
		log.Fatal(err)
	}

	err = db.ExampleData(pool)
	if err != nil {
		log.Fatal(err)
	}

	// f, err := db.GetFile(pool, ctx, "oldvid.mp4")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("file: %v\n", f)

	// uri := "I_Saw_The_TV_Glow_2024.mp4"
	// url, err := db.ResolveURI(pool, ctx, uri)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("Expanded %s to %s\n", uri, url)

	// newFile := shared.File{
	// 	BinId:           1,
	// 	Name:            "angry goose profile pic",
	// 	Hash:            "5f04192ad733bed11e6e391f9db836762015406209e3084846ff298f6a2fe7b4",
	// 	UploadTimestamp: time.Now().In(time.UTC),
	// }
	//
	//    db.CreateFile(pool, ctx, &newFile)
	//    if err != nil {
	//        log.Fatal(err)
	//    }

	// const PORT uint = 8080
	// mux := GetMux()
	// log.Printf("Listening on %d\n", PORT)
	// http.ListenAndServe(":"+fmt.Sprint(PORT), mux)
}
