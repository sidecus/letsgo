package raft

import (
	"errors"
)

var errorInvalidNodeCount = errors.New("node count is invalid, must be greater than 0")
var errorInvalidNodeID = errors.New("node id doesn't exist in the network")
var errorSendToSelf = errors.New("sending message to self is not allowed")
var errorInvalidMessage = errors.New("message is invalid")

// Network interfaces defines the interface for the underlying communication
type Network interface {
	Send(sourceNodeID int, targetNodID int, msg *Message) error
	Broadcast(sourceNodeID int, msg *Message) error
	GetChannel(nodeID int) (chan *Message, error)
}

// ChannelNetwork is a channel based message bus
type ChannelNetwork struct {
	nodeCnt  int
	channels []chan *Message
}

// CreateChannelNetwork creates a channel based network
func CreateChannelNetwork(n int) (*ChannelNetwork, error) {
	if n <= 0 {
		return nil, errorInvalidNodeCount
	}
	channels := make([]chan *Message, n)
	for i := range channels {
		// non buffered channel to mimic unrealiable network
		channels[i] = make(chan *Message)
	}

	return &ChannelNetwork{
		nodeCnt:  n,
		channels: channels,
	}, nil
}

// Send sends a message from the source node to target node
func (mb *ChannelNetwork) Send(sourceNodeID int, targetNodID int, msg *Message) error {
	switch {
	case sourceNodeID > mb.nodeCnt:
		return errorInvalidNodeID
	case targetNodID > mb.nodeCnt:
		return errorInvalidNodeID
	case sourceNodeID == targetNodID:
		return errorSendToSelf
	case msg == nil:
		return errorInvalidMessage
	}

	ch := mb.channels[targetNodID]
	// nonblocking lossy sending
	select {
	case ch <- msg:
	default:
	}

	return nil
}

// Broadcast a message to all other nodes
func (mb *ChannelNetwork) Broadcast(sourceNodeID int, msg *Message) error {
	switch {
	case sourceNodeID > mb.nodeCnt:
		return errorInvalidNodeID
	case msg == nil:
		return errorInvalidMessage
	}

	for i := range mb.channels {
		if i != sourceNodeID {
			mb.Send(sourceNodeID, i, msg)
		}
	}

	return nil
}

// GetChannel returns the receiving channel for the given node
func (mb *ChannelNetwork) GetChannel(nodeID int) (chan *Message, error) {
	if nodeID > mb.nodeCnt {
		return nil, errorInvalidNodeID
	}

	return mb.channels[nodeID], nil
}
