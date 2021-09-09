package engine

import (
	"database/sql"
	"fmt"
	"log"
)

func (n *Node) UpdateMetadata() {
	fmt.Println("Starting update of db metadata...")
	var err error
	for i := 0; ; i++ {
		nodes, err := n.DB.GetAll(100, i)
		if err != nil {
			break
		}
		for _, file := range nodes {
			if file.Type != "file" {
				continue
			}

			// Check for the size of a file if it isn't already known.
			if file.Size == 0 {
				file.Size, err = n.Node.GetSize(file.CID)
				if err != nil {
					continue
				}
			}

			// Check on the number of times a file is replicated across the cluster
			file.Replications, err = n.Node.FindProvs(file.CID, int(n.Manifest.Replications))
			if err != nil {
				continue
			}

			// Update the file entry in the database.
			n.DB.Update(file)
		}
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	fmt.Println("DB metadata update complete.")
}
