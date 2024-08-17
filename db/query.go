package db

import (
	"context"
	"database/sql"
	"errors"
	"file-cellar/storage"
	"time"
)

type DBCache struct {
	Bins    map[string]*storage.Bin
	Drivers map[string]storage.Driver
}

// Gets the full uri for a file given the file uri
// FIXME: doesn't resolve using Internal and External
func ResolveURI(ctx context.Context, db *sql.DB, uri string) (string, error) {
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

func GetFile(ctx context.Context, db *sql.DB, cache *DBCache, uri string) (*storage.File, error) {
	row := db.QueryRowContext(ctx, `
    SELECT files.name, files.hash, files.size, files.uploadTimestamp, bins.name
	FROM files
    INNER JOIN bins ON files.binID=bins.id
	WHERE files.relPath=?
	`, uri)

	f := new(storage.File)

	f.RelPath = uri
	var epochTime int64
	var binName string
	err := row.Scan(&f.Name, &f.Hash, &f.Size, &epochTime, &binName)

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

	bin, ok := cache.Bins[binName]
	if !ok {
		bin, err = GetBin(ctx, db, cache, binName)
		if err != nil {
			return nil, err
		}
	}
	f.Bin = bin

	return f, nil
}

func GetBin(ctx context.Context, db *sql.DB, cache *DBCache, binName string) (*storage.Bin, error) {
	bin, ok := cache.Bins[binName]
	if ok {
		return bin, nil
	} else {
		bin = new(storage.Bin)
	}

	row := db.QueryRowContext(ctx, `
    SELECT bins.id, bins.name, bins.externalURL, bins.internalURL, bins.redirect, drivers.name
    FROM bins
    INNER JOIN drivers ON bins.driverID=drivers.id
    WHERE bins.name=?`, binName)

	var driverName string
	err := row.Scan(&bin.Id, &bin.Name, &bin.Path.External, &bin.Path.Internal, &bin.Redirect, &driverName)
	if err != nil {
		return nil, err
	}

	driver, ok := cache.Drivers[driverName]
	if !ok {
		return nil, errors.New("can't find driver")
	}

	bin.Driver = driver
	cache.Bins[binName] = bin

	return bin, nil
}

func GetBins(ctx context.Context, db *sql.DB, cache *DBCache) error {
	cache.Bins = make(map[string]*storage.Bin)
	rows, err := db.QueryContext(ctx, `
    SELECT bins.id, bins.name, bins.internalURL, bins.externalURL, bins.redirect, drivers.name
    FROM bins
    INNER JOIN drivers ON bins.driverID = drivers.id`)
	if err == sql.ErrNoRows {
		logger.Printf("no bins found\n")
		return err
	}
	defer rows.Close()

	for rows.Next() {
		bin := new(storage.Bin)
		var driverName string
		err = rows.Scan(&bin.Name, &bin.Path.Internal, &bin.Path.External, &bin.Redirect, &driverName)

		if err != nil {
			logger.Printf("failed to read from database\n")
			continue
		}

		driver, ok := cache.Drivers[driverName]
		if !ok {
			logger.Printf("failed to find driver `%s` while querying bins\n", driverName)
			// TODO: create custom error and set it for return
			continue
		}
		bin.Driver = driver
		cache.Bins[bin.Name] = bin
	}

	return nil
}
