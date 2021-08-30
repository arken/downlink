package main

import (
	"log"
	"time"

	"github.com/arken/downlink/config"
	"github.com/arken/downlink/database"
	"github.com/arken/downlink/engine"
	"github.com/arken/downlink/ipfs"
	"github.com/arken/downlink/manifest"
	"github.com/arken/downlink/web"
	"github.com/go-co-op/gocron"
)

func main() {
	// Initialize the node's configuration
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Init(config.Global.Database.Path)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the node's manifest
	manifest, err := manifest.Init(
		config.Global.Manifest.Path,
		config.Global.Manifest.Url,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize embedded IPFS Node
	ipfs, err := ipfs.CreateNode(config.Global.Ipfs.Path, ipfs.NodeConfArgs{
		Addr:           config.Global.Ipfs.Addr,
		SwarmKey:       manifest.ClusterKey,
		BootstrapPeers: manifest.BootstrapPeers,
	})
	if err != nil {
		log.Fatal(err)
	}

	web := web.Node{
		DB:   db,
		Node: ipfs,
	}

	go web.Start(config.Global.Web.Addr, config.Global.Web.ForceHTTPS)

	// Initialize Arken Engine
	engine := engine.Node{
		Cfg:      &config.Global,
		DB:       db,
		Node:     ipfs,
		Manifest: manifest,
	}

	// Create Task Scheduler
	tasks := gocron.NewScheduler(time.UTC)

	// Configure Background Tasks
	// Check for and sync updates to the manifest every hour.
	tasks.Every(15).Minutes().Do(engine.SyncManifest)

	// Update file metadata (size & replications) daily
	tasks.Every(1).Day().Do(engine.UpdateMetadata)

	// Start Background Tasks
	tasks.StartBlocking()

}
