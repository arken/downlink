package database

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import sqlite driver for database interaction.
)

type DB struct {
	conn *sql.DB
	lock *sync.Mutex
}

type Node struct {
	Path         string
	Name         string
	Type         string
	Size         int64
	CID          string
	Parent       string
	Modified     time.Time
	Replications int
}

// Init opens and connects to the database.
func Init(path string) (result *DB, err error) {
	db, err := sql.Open("sqlite3", path+"?cache=shared&mode=memory")
	if err != nil {
		return nil, err
	}

	// Create the nodes table if is doesn't already exist.
	// This will also create the database if it doesn't exist.
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS nodes(
			path TEXT NOT NULL,
			name TEXT,
			type TEXT NOT NULL,
			size INT(11),
			cid TEXT,
			parent TEXT, 
			modified DATETIME,
			replications INT,
			PRIMARY KEY(id)
		);`,
	)
	if err != nil {
		return nil, err
	}

	result = &DB{
		conn: db,
		lock: &sync.Mutex{},
	}
	return result, nil
}
