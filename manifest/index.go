package manifest

import (
	"github.com/arken/downlink/database"
)

func (m *Manifest) Index(db *database.DB) (<-chan database.Node, error) {
	// Create results channel
	results := make(chan database.Node, 50)

	// Initialize go routine to handle indexing
	go func() {
		indexWorker(db, results)

		// Close results channel
		close(results)
	}()

	return results, nil
}

func indexWorker(db *database.DB, output chan<- database.Node) {

}
