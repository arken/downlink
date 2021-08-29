package database

import (
	"database/sql"
	"time"
)

// Add inserts a File entry into the database if it doesn't exist already.
func (db *DB) Add(input Node) (err error) {
	// Attempt to grab lock.
	db.lock.Lock()
	defer db.lock.Unlock()

	// Ping the DB and open a connection if necessary
	err = db.conn.Ping()
	if err != nil {
		return err
	}

	_, err = db.get(input.Path)
	if err != nil {
		if err == sql.ErrNoRows {
			err = db.insert(input)
		} else {
			return err
		}
	} else {
		err = db.update(input)
	}
	return err
}

// Insert adds a Node entry to the database.
func (db *DB) insert(entry Node) (err error) {
	stmt, err := db.conn.Prepare(
		`INSERT INTO nodes(
			path,
			name,
			type,
			size,
			cid,
			parent,
			modified,
			replications
		) VALUES(?,?,?,?,?,?,?,?);`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		entry.Path,
		entry.Name,
		entry.Type,
		entry.Size,
		entry.CID,
		entry.Parent,
		time.Now().UTC(),
		entry.Replications,
	)
	return err
}
