package db

import (
	"database/sql"
	"file-cellar/storage"
)

type Manager struct {
	db      *sql.DB
	connStr string
	Bins    map[int64]*storage.Bin
	Drivers map[string]storage.Driver
}

// Gets a database manager, reusing a connection pool if one exists
//
// Don't worry, using this function doesn't make you a Karen ;)
func GetManager(connStr string, pragmas map[string]string) (*Manager, error) {
	m := new(Manager)
	m.connStr = connStr
	m.Bins = make(map[int64]*storage.Bin)
	m.Drivers = make(map[string]storage.Driver)

	db, err := getPool(connStr, pragmas)
	if err != nil {
		return nil, err
	}

	m.db = db

	return m, nil
}

// Closes a manager's connection to the database connection pool.
func (m *Manager) Close() {
	closePool(m.connStr)
	m.db = nil
}
