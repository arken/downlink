package database

// Remove deletes and returns an entry from the database.
func (db *DB) Remove(path string) (result Node, err error) {
	// Attempt to grab lock.
	db.lock.Lock()
	defer db.lock.Unlock()

	// Ping the DB and open a connection if necessary
	err = db.conn.Ping()
	if err != nil {
		return result, err
	}

	// Get the current value of the entry in the DB before removing
	result, err = db.get(path)
	if err != nil {
		return result, err
	}

	// Remove the entry from the DB
	err = db.remove(path)
	return result, err
}

// remove deletes an entry to the DB.
func (db *DB) remove(path string) (err error) {
	stmt, err := db.conn.Prepare(
		"DELETE FROM nodes WHERE path = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path)
	return err
}
