package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var SQLITE_DEFAULT_PRAGMAS = map[string]string{
	"foreign_keys": "ON",
	"journal_mode": "wal",
	"synchronous":  "normal",
}

var logger *log.Logger

type dbPoolCounts struct {
	Alive uint32
	db    *sql.DB
}

// Sets pragmas for a database connection pool
//
// Returns a non nil value when an error occurs during SQL statement execution
func setPragmas(db *sql.DB, pragmas map[string]string) error {
	for k, v := range pragmas {
		_, err := db.Exec(fmt.Sprintf("PRAGMA %s = %s", k, v))
		if err != nil {
			logger.Printf("Error setting pragma %s to %s\n", k, v)
			return err
		}
	}

	return nil
}

// Gets a database connection pool
// Only sets pragmas for the connection pool on creation.
func getPool(connStr string, pragmas map[string]string) (*sql.DB, error) {
	pool, err := sql.Open("sqlite3", connStr)
	if err != nil {
		logger.Printf("Failed to open sqlite3 connection to %s", connStr)
		return nil, err
	}
	logger.Printf("Created sqlite3 connnection pool: %s", connStr)

	if err = setPragmas(pool, pragmas); err != nil {
		logger.Printf("Failed to set pragmas for %s", connStr)
		return nil, err
	}
	logger.Printf("Succesfully set pragmas for %s\n", connStr)

	return pool, nil
}

// Initializes tables in a database and sets indexes
//
// error is non nil if an error occurs while executing any SQL statement
func InitTables(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS serverConfig (
        key TEXT UNIQUE NOT NULL,
        value TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS drivers(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS bins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    driverID INTEGER,
    name TEXT UNIQUE NOT NULL,
    externalURL TEXT UNIQUE NOT NULL,
    internalURL TEXT NOT NULL,
    redirect INTEGER NOT NULL CHECK(redirect IN (0, 1)),
    FOREIGN KEY(driverID) REFERENCES drivers(id)
    )`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    binID INTEGER,
    name TEXT NOT NULL,
    hash TEXT NOT NULL,
    size INTEGER NOT NULL,
    relPath TEXT UNIQUE NOT NULL,
    uploadTimestamp INTEGER,
    FOREIGN KEY(binID) REFERENCES bins(id)
    )`)
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_files_date ON files(uploadTimestamp)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_files_name on files(name)")
	if err != nil {
		return err
	}

	logger.Println("Initialized Tables")
	return nil
}

func init() {
	logger = log.New(os.Stdout, "[DB]: ", log.LUTC|log.Ldate|log.Ltime)
}
