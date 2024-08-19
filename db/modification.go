package db

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"file-cellar/storage"
	"fmt"
	"strings"
)

// registers a storage driver
func (m *Manager) AddDriver(ctx context.Context, d storage.Driver) bool {
	result, err := m.db.ExecContext(ctx, `
    INSERT INTO drivers (name)
    VALUES (?)
    `, d.Name())

	if err != nil {
		logger.Print(err)
		return false
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Print(err)
	} else {
		d.SetId(id)
	}

	m.Drivers[d.Name()] = d

	return true
}

// adds a storage bin to the database and returns its index
func (m *Manager) AddBin(ctx context.Context, bin *storage.Bin, driverID int64) (int64, error) {
	result, err := m.db.ExecContext(ctx,
		`INSERT INTO bins (driverID, name, externalURL, internalURL, redirect)
        VALUES (?,?,?,?,?)`,
		driverID, bin.Name, bin.Path.External, bin.Path.Internal, bin.Redirect)
	if err != nil {
		logger.Print(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Print(err)
	}
	m.Bins[id] = bin

	return id, nil
}

// Assigns a relative path to a file
func (m *Manager) AddFile(ctx context.Context, f *storage.File) error {
	unixTime := f.UploadTimestamp.Unix()

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%d", f.Name, f.Hash, unixTime)))
	encoding := base64.URLEncoding.EncodeToString(hash[:])
	f.RelPath = strings.Trim(encoding, "=")

	_, err := m.db.ExecContext(ctx, `
    INSERT INTO files (binID, name, hash, size, relPath, uploadTimestamp)
    VALUES (?,?,?,?,?,?)`,
		f.Bin.Id, f.Name, f.Hash, f.Size, f.RelPath, unixTime)

	if err != nil {
		logger.Print(err)
	}

	return err
}

// Removes a file from the database
func (m *Manager) RemoveFile(ctx context.Context, uri string) (bool, error) {

	result, err := m.db.ExecContext(ctx, `
    DELETE FROM files
    WHERE relPath=?`, uri)

	if err != nil {
		logger.Printf("Failed to remove %s\n", uri)
		logger.Print(err)
		return false, err
	}

	count, err := result.RowsAffected()
	return count > 0, err
}
