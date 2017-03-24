package storage

import (
	"../../src"
	"../../src/logger"
	"fmt"
	"time"
)

type OutputEntry struct {
	Id                int    `json:'id'`
	HandlerIdentifier string `json:'handlerIdentifier'`
	CommandName       string
	Output            string `json:'output'`
	Timestamp         int32  `json:'timestamp'`
}

func GetUnsentEntries(maxEntries int) []OutputEntry {

	rows, err := DBConn.Query(
		"SELECT id, handler_identifier, output, timestamp FROM entries WHERE is_sent = 0 LIMIT ?",
		maxEntries,
	)
	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error parsing entries: %s", err),
		)
	}

	var entries []OutputEntry

	for rows.Next() {
		var entry OutputEntry
		err = rows.Scan(&entry.Id, &entry.HandlerIdentifier, &entry.Output, &entry.Timestamp)
		if err != nil {
			src.ExitApplicationWithMessage(
				fmt.Sprintf("Error parsing entries: %s", err),
			)
		}

		entries = append(entries, entry)
	}

	// fmt.Println(entries)

	return entries

}

func StoreOutputEntries(entries []OutputEntry) {
	if len(entries) == 0 {
		return
	}

	msStart := time.Now().UnixNano() / int64(time.Millisecond)

	tx, err := DBConn.Begin()

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error starting transaction: %s", err),
		)
	}

	stmt, err := tx.Prepare(`
    	INSERT INTO entries
    	(handler_identifier, command_name, output, timestamp) 
    	values(?, ?, ?, ?)
	`)

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error preparing statement: %s", err),
		)
	}

	for _, entry := range entries {
		_, err = stmt.Exec(
			entry.HandlerIdentifier,
			entry.CommandName,
			entry.Output,
			entry.Timestamp,
		)

		if err != nil {
			src.ExitApplicationWithMessage(
				fmt.Sprintf("Error inserting entry: %s", err),
			)
		}
	}

	tx.Commit()

	msEnd := time.Now().UnixNano() / int64(time.Millisecond)

	logger.Logger.Log(
		fmt.Sprintf(
			"Stored %d entries of %s in %dms\n",
			len(entries),
			entries[0].HandlerIdentifier,
			(msEnd - msStart),
		),
	)

}

func (self OutputEntry) Store() bool {

	stmt, err := DBConn.Prepare(`
    	INSERT INTO entries
    	(handler_identifier, command_name, output, timestamp) 
    	values(?, ?, ?, ?)
	`)

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error creating entry insert statement: %s", err),
		)
	}

	_, err = stmt.Exec(
		self.HandlerIdentifier,
		self.CommandName,
		self.Output,
		self.Timestamp,
	)

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error inserting entry: %s", err),
		)
	}

	return true
}
