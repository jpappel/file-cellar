package db

import (
	"database/sql"
	"file-cellar/storage"
)

var Managers map[string]*Manager

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
	if connStr != ":memory:" {
		if m, ok := Managers[connStr]; ok {
			return m, nil
		}
	}

	m := &Manager{
		connStr: connStr,
		Bins:    make(map[int64]*storage.Bin),
		Drivers: make(map[string]storage.Driver),
	}

	db, err := getPool(connStr, pragmas)
	if err != nil {
		return nil, err
	}

	m.db = db
	Managers[connStr] = m

	return m, nil
}

func (m *Manager) Init() error {
	// TODO: add field to avoid reinitialzing tables
	return InitTables(m.db)
}

// Closes a manager's database connection
func (m *Manager) Close() error {
	if err := m.db.Close(); err != nil {
		logger.Printf("Failed to close connection: %s\n%v", m.connStr, err)
		return err
	}
	m.db = nil

	return nil
}

func init() {
	Managers = make(map[string]*Manager)
}
