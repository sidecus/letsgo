package raft

import (
	"log"
	"sync"
)

// Cluster manages the raft cluster
type Cluster struct {
	size  int
	nodes []INode
}

// Start creates and starts a raft cluster with the given number of nodes on the network
func Start(size int, network Network, logger *log.Logger) {
	nodes := make([]INode, size)
	cluster := &Cluster{
		size:  size,
		nodes: nodes,
	}

	var wg sync.WaitGroup

	// Create nodes and start them
	for i := range nodes {
		wg.Add(1)
		nodes[i] = CreateNode(i, cluster, network, logger)
		nodes[i].Start()
	}

	wg.Wait()
}
