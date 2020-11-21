package raft

import (
	"log"
	"sync"
)

// LocalCluster manages the raft cluster on one host
type LocalCluster struct {
	size  int
	nodes []INode
}

// CreateLocalCluster creates a local cluster
func CreateLocalCluster(size int, network Network, logger *log.Logger) *LocalCluster {
	nodes := make([]INode, size)

	for i := range nodes {
		nodes[i] = CreateNode(i, size, network, logger)
	}

	return &LocalCluster{
		size:  size,
		nodes: nodes,
	}
}

// Start creates and starts a raft cluster with the given number of nodes on the network
func (cluster *LocalCluster) Start() {
	var wg sync.WaitGroup

	// Create nodes and start them
	for i := range cluster.nodes {
		wg.Add(1)
		go cluster.nodes[i].Start(&wg)
	}

	wg.Wait()
}
