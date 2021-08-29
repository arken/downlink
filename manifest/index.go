package manifest

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arken/downlink/database"
)

type Result struct {
	Node   database.Node
	Status string
}

func (m *Manifest) Index(db *database.DB) (<-chan Result, error) {
	// Create results channel
	results := make(chan Result, 50)

	// Initialize go routine to handle indexing
	go func() {
		m.indexWorker(db, results)

		// Close results channel
		close(results)
	}()
	return results, nil
}

func (m *Manifest) indexWorker(db *database.DB, output chan<- Result) {
	// Store current time to find files not touched during index
	start := time.Now().UTC()

	// Walk through entire repository directory structure to look for .ks files.
	err := filepath.Walk(m.path, func(path string, info os.FileInfo, err error) error {
		// On each interation of the "walk" this function will check if a manifest
		// file and parse for file IDs if true.
		if strings.HasSuffix(filepath.Base(path), ".ks") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			// Add filepaths all the way up to the root.
			fpath := strings.TrimSuffix(strings.TrimPrefix(path, m.path), ".ks")
			for {
				result := Result{
					Status: "add",
					Node: database.Node{
						Path:   fpath,
						Name:   filepath.Base(fpath),
						Type:   "dir",
						Parent: filepath.Dir(fpath),
					},
				}

				// Check if entry is already in the database.
				_, err := db.Get(fpath)
				if err == nil || err != sql.ErrNoRows {
					db.Update(result.Node)
				} else {
					output <- result
				}

				// Break the loop at the root node
				if fpath == "/" {
					break
				}

				// Climb up the file path
				fpath = filepath.Dir(fpath)
			}

			// Generate parent path for files in manifest file
			parentPath := strings.TrimSuffix(strings.TrimPrefix(path, m.path), ".ks")

			// Open the files for reading.
			scanner := bufio.NewScanner(file)

			// Scan one line at a time.
			for scanner.Scan() {
				// Split data on white space.
				data := strings.Fields(scanner.Text())

				result := Result{
					Status: "add",
					Node: database.Node{
						Path:   filepath.Join(parentPath, data[1]),
						Name:   data[1],
						Type:   "file",
						CID:    data[0],
						Parent: parentPath,
					},
				}

				// Check if entry is already in the database.
				_, err := db.Get(filepath.Join(parentPath, data[1]))
				if err == nil || err != sql.ErrNoRows {
					db.Update(result.Node)
					continue
				}

				output <- result
			}

			// Close the file after fully parsed.
			err = file.Close()
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	// Check for deleted files by looking for everything not touched by the index
	for i := 0; ; i++ {
		nodes, err := db.GetAllOlderThan(start, 100, i)
		if err != nil {
			break
		}
		for _, file := range nodes {
			output <- Result{
				Status: "remove",
				Node:   file,
			}
		}
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
}
