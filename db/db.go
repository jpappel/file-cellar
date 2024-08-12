package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbPools map[string]*sql.DB = make(map[string]*sql.DB)

var SQLITE_DEFAULT_PRAGMAS = map[string]string{
	"foreign_keys": "ON",
	"journal_mode": "wal",
	"synchronous":  "normal",
}

var logger = log.New(os.Stdout, "[DB]: ", log.LUTC|log.Ldate|log.Ltime)

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

// Gets a database connection pool, creating one if needed.
// Only sets pragmas for the connection pool on creation.
func GetPool(connStr string, pragmas map[string]string) (*sql.DB, error) {
	pool, ok := dbPools[connStr]
	if ok {
		return pool, nil
	}

	pool, err := sql.Open("sqlite3", connStr)
	if err != nil {
		logger.Printf("Failed to open sqlite3 connection to %s", connStr)
		return nil, err
	}
	logger.Printf("Created sqlite3 connnection pool: %s", connStr)

	err = setPragmas(pool, pragmas)
	if err != nil {
		logger.Printf("Failed to set pragmas for %s", connStr)
		return nil, err
	}
	logger.Printf("Succesfully set pragmas for %s\n", connStr)

	dbPools[connStr] = pool
	return pool, nil
}

// Close a database connection pool
func ClosePool(connStr string) (bool, error) {
	pool, ok := dbPools[connStr]
	if !ok {
		return false, nil
	}

	err := pool.Close()
	if err != nil {
		logger.Printf("Failed to close connection %s\n%v", connStr, err)
		return false, err
	}

	logger.Printf("Closed sqlite3 connection to %s\n", connStr)
	return true, nil
}

// Initializes tables in a database and sets indexes
//
// error is non nil if an error occurs while executing any SQL statement
func InitTables(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS serverConfig (
		key TEXT UNIQUE NOT NULL,
		value TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS drivers(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	// TODO: rename bins to something more semantic
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		driverID INTEGER,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		FOREIGN KEY(driverID) REFERENCES drivers(id)
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		binID INTEGER,
		name TEXT NOT NULL,
		hash TEXT NOT NULL,
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

func ExampleData(db *sql.DB) error {
	_, err := db.Exec(`
    INSERT INTO files (binID, name, hash, relPath, uploadTimestamp)
    VALUES
    (1, 'childhood video' , 'af8182a217f6c4ae4abb6d52951f6e7a2cac3a4d59889e4a7a3cce87ac0ae508' , 'oldvid.mp4' , 1000209017),
    (1, 'marriage photo', 'a0856e75fc1f1ec0d2fed17d534fbc1756770dbb0cc83788cbf8ca861c885fc0', 'WeddingAltar5.jpg', 451309817),
    (3, 'I saw the tv glow', '7b1a56dfcba8ce808cb6392e2403f895afb1f210b85b7d3ad324d365432f01fa', 'I_Saw_The_TV_Glow_2024.mp4', 1718538617),
    (2, 'dota2', '15c11ed3bd0eb92d6d54de44b36131643268e28f4aac9229f83231a0670c290c', 'Dota2Beta', 1373370617)
    `)
    return err
}

// func SetServerTable(db *sql.DB, params map[string]string) error
// func GetServerTable(db *sql.DB) (map[string]string, error)
