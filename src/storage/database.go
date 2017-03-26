package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DBConn *sql.DB
)

func OpenDatabase(dbName string) error {
	var err error
	DBConn, err = sql.Open("sqlite3", dbName)
	return err
}

func CreateStructure() error {
	_, err := DBConn.Exec(`
        CREATE TABLE IF NOT EXISTS 'entries' (
            'id' INTEGER PRIMARY KEY AUTOINCREMENT,
            'handler_identifier' VARCHAR(128) NOT NULL,
            'command_name' VARCHAR(128) NOT NULL,
            'output' TEXT NOT NULL,
            'is_sent' BOOLEAN DEFAULT(0),
            'timestamp' INT32 NOT NULL
        );
        
    `)

	if err != nil {
		return err
	}
	_, err = DBConn.Exec(`
        CREATE INDEX IF NOT EXISTS is_sent_idx ON entries (is_sent);
    `)

	if err != nil {
		return err
	}

	_, err = DBConn.Exec(`
        CREATE INDEX IF NOT EXISTS handler_identifier_idx ON entries (handler_identifier);
    `)

	return err
}
