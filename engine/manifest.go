package engine

import (
	"log"
	"os"

	"github.com/arken/downlink/manifest"
	"github.com/go-git/go-git/v5"
)

func (n *Node) SyncManifest() {
	// Pull changes from upstream manifest
	err := n.Manifest.Pull()
	if err != nil {
		// Check if a non-fast forward update occurred while pulling.
		if err == git.ErrNonFastForwardUpdate {
			// Remove manifest and re-clone.
			err = os.RemoveAll(n.Cfg.Manifest.Path)
			if err != nil {
				log.Println(err)
			}

			// Re-initialize the manifest
			n.Manifest, err = manifest.Init(
				n.Cfg.Manifest.Path,
				n.Cfg.Manifest.Url,
			)
		}
		if err != nil {
			log.Println(err)
		}
	}

	// Update manifest settings
	err = n.Manifest.Decode()
	if err != nil {
		log.Println(err)
	}

	// Index changes from manifest
	results, err := n.Manifest.Index(n.DB)
	if err != nil {
		log.Println(err)
	}

	for result := range results {
		// If file has been added to the manifest, add it
		// to the database.
		if result.Status == "add" {

			// Add file to database.
			err = n.DB.Add(result.Node)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		// If file has been deleted from the manifest, remove
		// it from the database.
		if result.Status == "remove" {
			_, err := n.DB.Remove(result.Node.Path)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
