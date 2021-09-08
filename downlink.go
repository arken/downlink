package main

import (
	"fmt"
	"log"
	"sync"
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
		Lock:     &sync.Mutex{},
	}

	fmt.Println("Setting up background task system...")

	// Create Task Scheduler
	tasks := gocron.NewScheduler(time.UTC)

	// Set the max number of concurrent jobs to 3.
	tasks.SetMaxConcurrentJobs(1, gocron.WaitMode)

	// Configure Background Tasks
	// Check for and sync updates to the manifest every hour.
	syncManifest, err := tasks.Every(15).Minutes().Do(engine.SyncManifest)
	if err != nil {
		log.Fatal(err)
	}
	syncManifest.SingletonMode()

	// Update file metadata (size & replications) daily
	updateMetadata, err := tasks.Every(1).Day().Do(engine.UpdateMetadata)
	if err != nil {
		log.Fatal(err)
	}
	updateMetadata.SingletonMode()

	fmt.Println("Background tasks setup.")

	// Start Background Tasks
	tasks.StartBlocking()

}
