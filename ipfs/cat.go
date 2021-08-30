package ipfs

import (
	files "github.com/ipfs/go-ipfs-files"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
)

// Read a file from IPFS without pinning
func (n *Node) Cat(hash string) (files.File, error) {
	// Construct IPFS CID
	path := icorepath.New("/ipfs/" + hash)

	// Pin file to local storage within IPFS
	node, err := n.api.Unixfs().Get(n.ctx, path)
	if err != nil {
		return nil, err
	}

	// Convert node into file
	file := files.ToFile(node)
	return file, nil
}
