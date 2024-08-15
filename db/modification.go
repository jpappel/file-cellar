package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"file-cellar/shared"
	"file-cellar/storage"
	"fmt"
	"strings"
)

// registers a storage driver
func AddDriver(db *sql.DB, ctx context.Context, driverName string) bool {
	row := db.QueryRowContext(ctx, `
    SELECT name FROM drivers
    WHERE name=?
    `, driverName)

	var name string
	err := row.Scan(&name)
	if err != nil || driverName == name {
		return false
	}

	_, err = db.ExecContext(ctx, `
    INSERT INTO drivers (name)
    VALUES (?)
    `, driverName)

	if err != nil {
		return false
	}

	return true
}

// adds a storage bin to the database and returns its index
func AddBin(db *sql.DB, ctx context.Context, bin storage.Bin) (int64, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO bins (driverID, name, url)
        VALUES (?,?,?)`, nil, bin.Name, bin.Url) // FIXME: use driver id
	if err != nil {
		logger.Print(err)
	}

	return result.LastInsertId()
}

// Assigns a relative path to a file
func AddFile(db *sql.DB, ctx context.Context, f *shared.File) error {
	unixTime := f.UploadTimestamp.Unix()

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%d", f.Name, f.Hash, unixTime)))
	encoding := base64.URLEncoding.EncodeToString(hash[:])
	f.RelPath = strings.Trim(encoding, "=")

	_, err := db.ExecContext(ctx,
		`INSERT INTO files (binID, name, hash, size, relPath, uploadTimestamp)
    VALUES (?,?,?,?,?,?)`,
		f.BinId, f.Name, f.Hash, f.Size, f.RelPath, unixTime)

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
