package main

import (
	"fmt"
	"github.com/sidecus/letsgo/raft"
	"log"
	"runtime"
)

// TODO[sidecus]: switch to log.Default when new Go version is out
var logger = log.New(log.Writer(), log.Prefix(), log.Flags())

func raftDemo() {
	var clusterSize = runtime.NumCPU()

	fmt.Printf("Starting a %d node raft cluster\n", clusterSize)

	network, _ := raft.CreateChannelNetwork(clusterSize)
	raft.StartCluster(clusterSize, network, logger)
}
