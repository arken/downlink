package database

import (
	"database/sql"
	"time"
)

// Get searches for and returns a the corresponding entry from the
// database if the entry exists.
func (db *DB) Get(id string) (result Node, err error) {
	// Attempt to grab lock.
	db.lock.Lock()
	defer db.lock.Unlock()

	// Ping the DB and open a connection if necessary
	err = db.conn.Ping()
	if err != nil {
		return result, err
	}

	// Get and return entry from DB if it exists
	return db.get(id)
}

// get returns the matching entry from the db if it exists.
func (db *DB) get(path string) (result Node, err error) {
	row, err := db.conn.Query("SELECT * FROM nodes WHERE path = ?", path)
	if err != nil {
		return result, err
	}
	defer row.Close()
	if !row.Next() {
		return result, sql.ErrNoRows
	}
	err = row.Scan(
		&result.Path,
		&result.Name,
		&result.Type,
		&result.Size,
		&result.CID,
		&result.Parent,
		&result.Modified,
		&result.Replications,
	)

	return result, err
}

func (db *DB) GetChildren(parentPath string, limit, page int) (result []Node, err error) {
	// Attempt to grab lock.
	db.lock.Lock()
	defer db.lock.Unlock()

	// Create files slice with limit as size.
	result = []Node{}

	// Ping database to check that it still exists.
	err = db.conn.Ping()
	if err != nil {
		return result, err
	}

	rows, err := db.conn.Query(
		"SELECT * FROM nodes WHERE parent = ? LIMIT ? OFFSET ?;",
		parentPath,
		limit,
		limit*page,
	)
	if err != nil {
		return result, err
	}

	// Iterate through rows found and insert them into the list.
	for rows.Next() {
		var f Node

		err = rows.Scan(
			&f.Path,
			&f.Name,
			&f.Type,
			&f.Size,
			&f.CID,
			&f.Parent,
			&f.Modified,
			&f.Replications)

		if err != nil {
			rows.Close()
			return nil, err
		}

		result = append(result, f)
	}

	// Check for errors and return
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	if len(result) <= 0 {
		return nil, sql.ErrNoRows
	}

	return result, err
}

func (db *DB) GetAllOlderThan(age time.Time, limit, page int) (result []Node, err error) {
	// Attempt to grab lock.
	db.lock.Lock()
	defer db.lock.Unlock()

	// Create files slice with limit as size.
	result = []Node{}

	// Ping database to check that it still exists.
	err = db.conn.Ping()
	if err != nil {
		return result, err
	}

	rows, err := db.conn.Query(
		"SELECT * FROM nodes WHERE modified < ? LIMIT ? OFFSET ?;",
		age,
		limit,
		limit*page,
	)
	if err != nil {
		return result, err
	}

	// Iterate through rows found and insert them into the list.
	for rows.Next() {
		var f Node

		err = rows.Scan(
			&f.Path,
			&f.Name,
			&f.Type,
			&f.Size,
			&f.CID,
			&f.Parent,
			&f.Modified,
			&f.Replications)

		if err != nil {
			rows.Close()
			return nil, err
		}

		result = append(result, f)
	}

	// Check for errors and return
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	if len(result) <= 0 {
		return nil, sql.ErrNoRows
	}

	return result, err
}
