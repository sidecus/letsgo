package raft

import (
	"errors"
)

// Network interfaces defines the interface for the underlying communication among nodes
type Network interface {
	Send(sourceNodeID int, targetNodID int, msg *raftMessage) error
	Broadcast(sourceNodeID int, msg *raftMessage) error
	GetRecvChannel(nodeID int) (chan *raftMessage, error)
}

var errorInvalidNodeCount = errors.New("node count is invalid, must be greater than 0")
var errorInvalidNodeID = errors.New("node id doesn't exist in the network")
var errorSendToSelf = errors.New("sending message to self is not allowed")
var errorInvalidMessage = errors.New("message is invalid")

// boradcastAddress is a special NodeId representing broadcasting to all other nodes
const boradcastAddress = -1

// dummyNetworkRequest request type used by channelNetwork internally to send raftMessage to nodes
type dummyNetworkRequest struct {
	sender   int
	receiver int
	message  *raftMessage
}

// dummyNetwork is a channel based network implementation without real RPC calls
type dummyNetwork struct {
	size         int
	chSend       chan dummyNetworkRequest
	nodeChannels []chan *raftMessage
}

// CreateDummyNetwork creates a channel based network
func CreateDummyNetwork(n int) (Network, error) {
	if n <= 0 || n > 1024 {
		return nil, errorInvalidNodeCount
	}

	chSend := make(chan dummyNetworkRequest, 100)
	nodeChannels := make([]chan *raftMessage, n)
	for i := range nodeChannels {
		// non buffered channel to mimic unrealiable network
		nodeChannels[i] = make(chan *raftMessage)
	}

	network := &dummyNetwork{
		size:         n,
		chSend:       chSend,
		nodeChannels: nodeChannels,
	}

	// start the network
	go network.start()

	return network, nil
}

// sendToNode sends one message to one receiver
func (net *dummyNetwork) sendToNode(receiver int, msg *raftMessage) {
	// nonblocking lossy sending using channel
	ch := net.nodeChannels[receiver]
	select {
	case ch <- msg:
	default:
	}
}

// start starts the dummy network, monitoring for send requests and dispatch it to nodes
func (net *dummyNetwork) start() {
	for r := range net.chSend {
		if r.receiver != boradcastAddress {
			// single receiver
			net.sendToNode(r.receiver, r.message)
		} else {
			// broadcast to others
			for i := range net.nodeChannels {
				if i != r.sender {
					net.sendToNode(i, r.message)
				}
			}
		}
	}
}

// Send sends a message from the source node to target node
func (net *dummyNetwork) Send(sourceNodeID int, targetNodeID int, msg *raftMessage) error {
	switch {
	case sourceNodeID > net.size:
		return errorInvalidNodeID
	case targetNodeID > net.size:
		return errorInvalidNodeID
	case sourceNodeID == targetNodeID:
		return errorSendToSelf
	case msg == nil:
		return errorInvalidMessage
	}

	req := dummyNetworkRequest{
		sender:   sourceNodeID,
		receiver: targetNodeID,
		message:  msg,
	}

	net.chSend <- req

	return nil
}

// Broadcast a message to all other nodes
func (net *dummyNetwork) Broadcast(sourceNodeID int, msg *raftMessage) error {
	switch {
	case sourceNodeID > net.size:
		return errorInvalidNodeID
	case msg == nil:
		return errorInvalidMessage
	}

	req := dummyNetworkRequest{
		sender:   sourceNodeID,
		receiver: boradcastAddress,
		message:  msg,
	}

	net.chSend <- req

	return nil
}

// GetRecvChannel returns the receiving channel for the given node
func (net *dummyNetwork) GetRecvChannel(nodeID int) (chan *raftMessage, error) {
	if nodeID > net.size {
		return nil, errorInvalidNodeID
	}

	return net.nodeChannels[nodeID], nil
}
