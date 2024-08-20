package config

import "file-cellar/db"

var Server map[string]string
var DB_PRAGMAS map[string]string

func init() {
	DB_PRAGMAS = db.SQLITE_DEFAULT_PRAGMAS
	Server = make(map[string]string)
	Server["DBURL"] = "testing.db"
}
