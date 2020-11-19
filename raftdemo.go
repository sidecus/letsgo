package main

import (
	"fmt"
	"github.com/sidecus/letsgo/raft"
	"log"
)

const clusterSize = 5

// TODO[sidecus]: switch to log.Default when new Go version is out
var logger = log.New(log.Writer(), log.Prefix(), log.Flags())

func raftDemo() {
	fmt.Printf("Starting a %d node raft cluster\n", clusterSize)
	network, _ := raft.CreateChannelNetwork(clusterSize)
	raft.StartCluster(clusterSize, network, logger)
}
