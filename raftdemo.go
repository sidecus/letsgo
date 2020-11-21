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

	// Create and start a local channel based dummy network
	network, _ := raft.CreateChannelNetwork(clusterSize)

	// start a local cluster
	fmt.Printf("Starting a %d node raft cluster\n", clusterSize)
	cluster := raft.CreateLocalCluster(clusterSize, network, logger)
	cluster.Start()
}
