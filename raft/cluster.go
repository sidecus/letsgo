package raft

import (
	"log"
	"sync"
)

// Cluster manages the raft cluster
type Cluster struct {
	size    int
	nodes   []INode
	network Network
	logger  *log.Logger
}

// StartCluster creates and starts a raft cluster with the given number of nodes on the network
func StartCluster(size int, network Network, logger *log.Logger) {
	nodes := make([]INode, size)
	cluster := &Cluster{
		size:    size,
		nodes:   nodes,
		network: network,
		logger:  logger,
	}

	var wg sync.WaitGroup

	// Create nodes and start them
	for i := range nodes {
		recvCh, _ := network.GetChannel(i)
		nodes[i] = CreateNode(cluster, i, recvCh)

		wg.Add(1)
		nodes[i].Start()
	}

	wg.Wait()
}
