package db

import (
	"context"
	"database/sql"
	"file-cellar/drivers"
	"file-cellar/shared"
	"time"
)

// Gets the full uri for a file given the file uri
func ResolveURI(db *sql.DB, ctx context.Context, uri string) (string, error) {
	row := db.QueryRowContext(ctx,
		`SELECT concat(url, '/' ,relPath)
    FROM files
    INNER JOIN bins
    ON files.binID = bins.id
    WHERE relPath=?`, uri)

	url := ""
	err := row.Err()
	switch {
	case err == sql.ErrNoRows:
		logger.Printf("no file with uri%s\n", uri)
	case err != nil:
		logger.Printf("failure when querying for file %s\n%v", uri, err)
	default:
		row.Scan(&url)
	}

	return url, err
}

func GetFile(db *sql.DB, ctx context.Context, uri string) (shared.File, error) {
	row := db.QueryRowContext(ctx,
		`SELECT name, hash, uploadTimestamp, binID
	FROM files
	WHERE relPath=?
	`, uri)

	f := shared.File{RelPath: uri}
	var epochTime int64
	err := row.Scan(&f.Name, &f.Hash, &epochTime, &f.BinId)

	switch {
	case err == sql.ErrNoRows:
		logger.Printf("no file with uri %s\n", uri)
		return f, err
	case err != nil:
		logger.Printf("failure when querying for file %s\n%v", uri, err)
		return f, err
	default:
		f.UploadTimestamp = time.Unix(epochTime, 0)
	}

	return f, nil
}

func GetBins(db *sql.DB, ctx context.Context, drivers map[string]*drivers.Driver) (map[int]shared.Bin, error) {
	bins := make(map[int]shared.Bin)
	rows, err := db.QueryContext(ctx,
		`SELECT bins.id, bins.name, url, drivers.name
    FROM bins
    INNER JOIN drivers ON bins.driverID = drivers.id`)
	if err == sql.ErrNoRows {
		logger.Printf("no bins found\n")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var driverName string
		var bin shared.Bin
		rows.Scan(&id, &bin.Name, &bin.Url, &driverName)

		driver, ok := drivers[driverName]
		if !ok {
			logger.Printf("failed to find driver `%s` while querying bins\n", driverName)
			// TODO: create custom error and set it for return
			continue
		}
		bin.Driver = driver

		bins[id] = bin
	}

	return bins, nil
}
