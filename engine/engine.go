package engine

import (
	"sync"

	"github.com/arken/downlink/config"
	"github.com/arken/downlink/database"
	"github.com/arken/downlink/ipfs"
	"github.com/arken/downlink/manifest"
)

type Node struct {
	Cfg      *config.Config
	DB       *database.DB
	Node     *ipfs.Node
	Manifest *manifest.Manifest
	Lock     *sync.Mutex
}
