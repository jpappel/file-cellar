package db

import (
	"context"
	"database/sql"
	"errors"
	"file-cellar/storage"
	"time"
)

// Gets the full url for a file given the file path
func (m *Manager) Resolve(ctx context.Context, path string) (string, error) {
	row := m.db.QueryRowContext(ctx, `
    SELECT concat(internalURL, '/' ,relPath)
    FROM files
    INNER JOIN bins
    ON files.binID = bins.id
    WHERE relPath=?`, path)

	url := ""
	err := row.Err()
	switch {
	case err == sql.ErrNoRows:
		logger.Printf("no file with uri%s\n", path)
	case err != nil:
		logger.Printf("failure when querying for file %s\n%v", path, err)
	default:
		row.Scan(&url)
	}

	return url, err
}

func (m *Manager) GetFile(ctx context.Context, uri string) (*storage.File, error) {
	row := m.db.QueryRowContext(ctx, `
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

	bin, ok := m.Bins[binName]
	if !ok {
		bin, err = m.GetBin(ctx, binName)
		if err != nil {
			return nil, err
		}
	}
	f.Bin = bin

	return f, nil
}

func (m *Manager) GetBin(ctx context.Context, binName string) (*storage.Bin, error) {
	bin, ok := m.Bins[binName]
	if ok {
		return bin, nil
	} else {
		bin = new(storage.Bin)
	}

	row := m.db.QueryRowContext(ctx, `
    SELECT bins.id, bins.name, bins.externalURL, bins.internalURL, bins.redirect, drivers.name
    FROM bins
    INNER JOIN drivers ON bins.driverID=drivers.id
    WHERE bins.name=?`, binName)

	var driverName string
	err := row.Scan(&bin.Id, &bin.Name, &bin.Path.External, &bin.Path.Internal, &bin.Redirect, &driverName)
	if err != nil {
		return nil, err
	}

	driver, ok := m.Drivers[driverName]
	if !ok {
		return nil, errors.New("can't find driver")
	}

	bin.Driver = driver
	m.Bins[binName] = bin

	return bin, nil
}

func (m *Manager) GetBins(ctx context.Context) error {
	m.Bins = make(map[string]*storage.Bin)
	rows, err := m.db.QueryContext(ctx, `
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

		driver, ok := m.Drivers[driverName]
		if !ok {
			logger.Printf("failed to find driver `%s` while querying bins\n", driverName)
			// TODO: create custom error and set it for return
			continue
		}
		bin.Driver = driver
		m.Bins[bin.Name] = bin
	}

	return nil
}
