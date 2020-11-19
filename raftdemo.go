package main

import (
	"fmt"
	"github.com/sidecus/letsgo/raft"
)

const clusterSize = 5

func raftDemo() {
	fmt.Printf("Starting a %d node raft cluster\n", clusterSize)
	network, _ := raft.CreateChannelNetwork(clusterSize)
	raft.StartCluster(clusterSize, network)
}
