package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/sidecus/letsgo/raft"
)

func raftDemo() {
	var clusterSize = runtime.NumCPU() - 1

	// TODO[sidecus]: switch to log.Default when new Go version is out
	logger := log.New(log.Writer(), log.Prefix(), log.Flags())
	network, _ := raft.CreateDummyNetwork(clusterSize)

	fmt.Printf("Starting a %d node raft cluster\n", clusterSize)
	raft.Start(clusterSize, network, logger)
}
