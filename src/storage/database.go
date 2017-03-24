package storage

import (
	"../../src"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DBConn *sql.DB
)

func OpenDatabase(dbName string) {
	var err error
	DBConn, err = sql.Open("sqlite3", dbName)
	if err != nil {
		// Command not found for the handler, error and exit
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Issue with db: %s", err),
		)
	}
}

func CreateStructure() {
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
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error creating the table: %s", err),
		)
	}
	_, err = DBConn.Exec(`
        CREATE INDEX IF NOT EXISTS is_sent_idx ON entries (is_sent);
    `)

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error adding index: %s", err),
		)
	}
}
