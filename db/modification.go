package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"file-cellar/storage"
	"fmt"
	"strings"
)

// registers a storage driver
func AddDriver(ctx context.Context, db *sql.DB, driverName string) bool {
	_, err := db.ExecContext(ctx, `
    INSERT INTO drivers (name)
    VALUES (?)
    `, driverName)

	if err != nil {
		logger.Print(err)
		return false
	}

	return true
}

// adds a storage bin to the database and returns its index
func AddBin(ctx context.Context, db *sql.DB, bin storage.Bin, driverID int64) (int64, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO bins (driverID, name, externalURL, internalURL, redirect)
        VALUES (?,?,?,?,?)`,
		driverID, bin.Name, bin.Path.External, bin.Path.Internal, bin.Redirect)
	if err != nil {
		logger.Print(err)
	}

	return result.LastInsertId()
}

// Assigns a relative path to a file
func AddFile(ctx context.Context, db *sql.DB, f *storage.File) error {
	unixTime := f.UploadTimestamp.Unix()

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%d", f.Name, f.Hash, unixTime)))
	encoding := base64.URLEncoding.EncodeToString(hash[:])
	f.RelPath = strings.Trim(encoding, "=")

	_, err := db.ExecContext(ctx, `
    INSERT INTO files (binID, name, hash, size, relPath, uploadTimestamp)
    VALUES (?,?,?,?,?,?)`,
		f.Bin.Id, f.Name, f.Hash, f.Size, f.RelPath, unixTime)

	if err != nil {
		logger.Print(err)
	}

	return err
}

// Removes a file from the database
func RemoveFile(db *sql.DB, ctx context.Context, uri string) (bool, error) {

	result, err := db.ExecContext(ctx,
		`DELETE FROM files
    WHERE relPath=?`, uri)

	if err != nil {
		logger.Printf("Failed to remove %s\n", uri)
		logger.Print(err)
		return false, err
	}

	count, err := result.RowsAffected()
	return count > 0, err
}
