package raft

import (
	"errors"
)

// INetwork interfaces defines the interface for the underlying communication among nodes
type INetwork interface {
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

// channelNetworkReq request type used by channelNetworkReq internally to send raftMessage to nodes
type channelNetworkReq struct {
	sender   int
	receiver int
	message  *raftMessage
}

// channelNetwork is a channel based network implementation without real RPC calls
type channelNetwork struct {
	size  int
	cin   chan channelNetworkReq
	couts []chan *raftMessage
}

// CreateChannelNetwork creates a channelNetwork (local machine channel based network)
// and starts it. It mimics real network behavior by retrieving requests from cin and dispatch to couts
func CreateChannelNetwork(n int) (INetwork, error) {
	if n <= 0 || n > 1024 {
		return nil, errorInvalidNodeCount
	}

	cin := make(chan channelNetworkReq, 100)
	couts := make([]chan *raftMessage, n)
	for i := range couts {
		// non buffered channel to mimic unrealiable network
		couts[i] = make(chan *raftMessage)
	}

	net := &channelNetwork{
		size:  n,
		cin:   cin,
		couts: couts,
	}

	// start the network, which reads from cin and dispatches to one or more couts
	go func() {
		for r := range net.cin {
			if r.receiver != boradcastAddress {
				net.sendToNode(r.receiver, r.message)
			} else {
				// broadcast to others
				for i := range net.couts {
					if i != r.sender {
						net.sendToNode(i, r.message)
					}
				}
			}
		}
	}()

	return net, nil
}

// sendToNode sends one message to one receiver
func (net *channelNetwork) sendToNode(receiver int, msg *raftMessage) {
	// nonblocking lossy sending using channel
	ch := net.couts[receiver]
	select {
	case ch <- msg:
	default:
	}
}

// Send sends a message from the source node to target node
func (net *channelNetwork) Send(sourceNodeID int, targetNodeID int, msg *raftMessage) error {
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

	req := channelNetworkReq{
		sender:   sourceNodeID,
		receiver: targetNodeID,
		message:  msg,
	}

	net.cin <- req

	return nil
}

// Broadcast a message to all other nodes
func (net *channelNetwork) Broadcast(sourceNodeID int, msg *raftMessage) error {
	switch {
	case sourceNodeID > net.size:
		return errorInvalidNodeID
	case msg == nil:
		return errorInvalidMessage
	}

	req := channelNetworkReq{
		sender:   sourceNodeID,
		receiver: boradcastAddress,
		message:  msg,
	}

	net.cin <- req

	return nil
}

// GetRecvChannel returns the receiving channel for the given node
func (net *channelNetwork) GetRecvChannel(nodeID int) (chan *raftMessage, error) {
	if nodeID > net.size {
		return nil, errorInvalidNodeID
	}

	return net.couts[nodeID], nil
}
