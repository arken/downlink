package database

import "time"

// Update attempts to modify an existing entry in the database.
func (db *DB) Update(entry Node) (old Node, err error) {
	// Attempt to grab lock.
	db.lock.Lock()
	defer db.lock.Unlock()

	// Ping the DB and open a connection if necessary
	err = db.conn.Ping()
	if err != nil {
		return old, err
	}

	// Check for entry in DB
	old, err = db.get(entry.Path)
	if err != nil {
		return old, err
	}

	// Update the entry if it exists.
	err = db.update(entry)
	return old, err
}

// update changes a file's status in the database.
func (db *DB) update(entry Node) (err error) {
	stmt, err := db.conn.Prepare(
		`UPDATE nodes SET
			name = ?,
			type = ?,
			size = ?,
			cid = ?,
			parent = ?,
			modified = ?,
			replications = ?,
			WHERE id = ?;`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		entry.Name,
		entry.Type,
		entry.Size,
		entry.CID,
		entry.Parent,
		time.Now().UTC(),
		entry.Replications,
		entry.Path,
	)
	return err
}
