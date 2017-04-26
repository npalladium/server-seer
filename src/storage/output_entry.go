package storage

import (
	"../../src/logger"
	"fmt"
	"strings"
	"time"
)

type OutputEntry struct {
	Id                int    `json:'id'`
	HandlerIdentifier string `json:'handlerIdentifier'`
	CommandName       string
	Output            string `json:'output'`
	Timestamp         int32  `json:'timestamp'`
}

func GetUnsentEntries(maxEntries int) ([]OutputEntry, error) {

	rows, err := DBConn.Query(
		"SELECT id, handler_identifier, output, timestamp FROM entries WHERE is_sent = 0 LIMIT ?",
		maxEntries,
	)
	if err != nil {
		return nil, err
	}

	var entries []OutputEntry

	for rows.Next() {
		var entry OutputEntry
		err = rows.Scan(&entry.Id, &entry.HandlerIdentifier, &entry.Output, &entry.Timestamp)

		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil

}

func MarkEntriesSent(entries []OutputEntry) {
	var ids []int
	for _, entry := range entries {
		ids = append(ids, entry.Id)
	}

	str := fmt.Sprintf("UPDATE entries SET is_sent = 1 WHERE id IN (%s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]"))

	_, err := DBConn.Exec(
		str,
	)

	if err != nil {
		panic(err)
	}
}

func StoreOutputEntries(entries []OutputEntry) error {
	if len(entries) == 0 {
		return nil
	}

	msStart := time.Now().UnixNano() / int64(time.Millisecond)

	tx, err := DBConn.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
    	INSERT INTO entries
    	(handler_identifier, command_name, output, timestamp) 
    	values(?, ?, ?, ?)
	`)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		_, err = stmt.Exec(
			entry.HandlerIdentifier,
			entry.CommandName,
			entry.Output,
			entry.Timestamp,
		)

		if err != nil {
			return err
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

	return nil

}

func (self OutputEntry) Store() error {

	stmt, err := DBConn.Prepare(`
    	INSERT INTO entries
    	(handler_identifier, command_name, output, timestamp) 
    	values(?, ?, ?, ?)
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		self.HandlerIdentifier,
		self.CommandName,
		self.Output,
		self.Timestamp,
	)

	if err != nil {
		return err
	}

	return nil
}

func DeleteOldEntries(oldestEntrySeconds int) error {
	timestampOldest := time.Now().Unix() - int64(oldestEntrySeconds)
	str := fmt.Sprintf("DELETE FROM entries WHERE timestamp < %d", timestampOldest)

	_, err := DBConn.Exec(
		str,
	)

	return err
}
