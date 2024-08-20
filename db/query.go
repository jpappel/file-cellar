package db

import (
	"context"
	"database/sql"
	"errors"
	"file-cellar/storage"
	"fmt"
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
	err := row.Scan(&url)
	if err == sql.ErrNoRows {
		logger.Printf("no file with uri: %s\n", path)
		return "", err
	}
	if err != nil {
		logger.Printf("failure when querying for file %s\n%v", path, err)
		return "", err
	}

	return url, err
}

func (m *Manager) GetFile(ctx context.Context, uri string) (*storage.FileInfo, error) {
	row := m.db.QueryRowContext(ctx, `
    SELECT binID, name, hash, size, uploadTimestamp
    FROM files
    WHERE files.relPath=?
	`, uri)

	f := new(storage.FileInfo)

	f.RelPath = uri
	var epochTime int64
	var binId int64
	err := row.Scan(&binId, &f.Name, &f.Hash, &f.Size, &epochTime)

	switch {
	case err == sql.ErrNoRows:
		logger.Printf("no file with uri %s\n", uri)
		return nil, err
	case err != nil:
		logger.Printf("failure when querying for file %s\n%v", uri, err)
		return nil, err
	default:
		f.UploadTimestamp = time.Unix(epochTime, 0)
	}

	bin, ok := m.Bins[binId]
	if !ok {
		bin, err = m.GetBin(ctx, binId)
		if err != nil {
			return nil, err
		}
	}
	f.Bin = bin

	return f, nil
}

func (m *Manager) GetBin(ctx context.Context, id int64) (*storage.Bin, error) {
	bin, ok := m.Bins[id]
	if ok {
		return bin, nil
	}
	bin = new(storage.Bin)
	bin.Id = id

	row := m.db.QueryRowContext(ctx, `
    SELECT bins.name, bins.externalURL, bins.internalURL, bins.redirect, drivers.name
    FROM bins
    INNER JOIN drivers ON bins.driverID=drivers.id
    WHERE bins.id=?`, id)

	var driverName string
	err := row.Scan(&bin.Name, &bin.Path.External, &bin.Path.Internal, &bin.Redirect, &driverName)
	if err != nil {
		fmt.Println("error after scan: ", err)
		return nil, err
	}

	driver, ok := m.Drivers[driverName]
	if !ok {
		driver, err = m.GetDriver(ctx, driverName)
		if err != nil {
			return nil, err
		}
	}

	bin.Driver = driver
	m.Bins[id] = bin

	return bin, nil
}

func (m *Manager) GetDriver(ctx context.Context, driverName string) (storage.Driver, error) {
	row := m.db.QueryRowContext(ctx, `
    SELECT id
    FROM drivers
    WHERE name=?
    `, driverName)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		return nil, err
	}

	var driver storage.Driver

	switch driverName {
	case "LocalDriver":
		driver = storage.NewLocalDriver()
		driver.SetId(id)
	default:
		return nil, errors.New("unknown driver")
	}

	return driver, nil
}

// Clear a managers bins and recreates them according to the database
func (m *Manager) GetBins(ctx context.Context) error {
	m.Bins = make(map[int64]*storage.Bin)
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
		var id int64
		var driverName string
		err = rows.Scan(&id, &bin.Name, &bin.Path.Internal, &bin.Path.External, &bin.Redirect, &driverName)

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
		m.Bins[id] = bin
	}

	return nil
}
