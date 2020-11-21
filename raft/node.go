package raft

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// NodeState - node state type, can be one of 3: follower, candidate or leader
type NodeState int

// NodeState allowed values
const (
	follower  = 0
	candidate = 1
	leader    = 2
)

// INode defines the interface for a node
type INode interface {
	State() NodeState
	SetState(newState NodeState) NodeState

	StopElectionTimer()
	ResetElectionTimer()
	StopHeartbeatTimer()
	ResetHeartbeatTimer()

	StartElection() bool
	Vote(electMsg *raftMessage) bool
	CountVotes(ballotMsg *raftMessage) bool
	AckHeartbeat(hbMsg *raftMessage) bool
	SendHeartbeat() bool

	Start(wg *sync.WaitGroup)
}

const minElectionTimeoutMs = 3500                     // millisecond
const maxElectionTimeoutMs = 5000                     // millisecond
const heartbeatTimeoutMs = minElectionTimeoutMs + 100 // millisecond, larger value so that we can mimic some failures

// Node represents a raft node
type Node struct {
	id            int
	term          int
	state         NodeState
	lastVotedTerm int
	votes         []bool
	size          int

	electionTimer  *time.Timer // timer for election timeout, used by follower and candidate
	heartbeatTimer *time.Timer // timer for heartbeat, used by leader

	sm      nodeStateMachine
	network Network // underlying network implementation for sending/receiving messages
	logger  *log.Logger
}

// CreateNode creates a new raft node
func CreateNode(id int, size int, network Network, logger *log.Logger) INode {
	// Initialize timer objects (stopped immediately)
	electionTimer := time.NewTimer(time.Hour)
	electionTimer.Stop()
	heartbeatTimer := time.NewTimer(time.Hour)
	heartbeatTimer.Stop()

	return &Node{
		id:            id,
		term:          0,
		state:         follower,
		lastVotedTerm: 0,
		votes:         make([]bool, size),
		size:          size,

		electionTimer:  electionTimer,
		heartbeatTimer: heartbeatTimer,

		sm:      raftNodeSM,
		network: network,
		logger:  logger,
	}
}

// getElectionTimeout gets a new random election timeout
func (node *Node) getElectionTimeout() time.Duration {
	timeout := rand.Intn(maxElectionTimeoutMs-minElectionTimeoutMs) + minElectionTimeoutMs
	return time.Duration(timeout) * time.Millisecond
}

// processMessage passes the message through the node statemachine
// it returns a signal about whether we should quit
func (node *Node) processMessage(msg *raftMessage) bool {
	node.sm.ProcessMessage(node, msg)
	return false
}

// State returns the node's current state
func (node *Node) State() NodeState {
	return node.state
}

// SetState changes the node's state
func (node *Node) SetState(newState NodeState) NodeState {
	oldState := node.state
	node.state = newState
	return oldState
}

// StopElectionTimer stops the election timer for the node
func (node *Node) StopElectionTimer() {
	stopTimer(node.electionTimer)
}

// ResetElectionTimer resets the election timer for the node
func (node *Node) ResetElectionTimer() {
	stopTimer(node.electionTimer)
	node.electionTimer.Reset(node.getElectionTimeout())
}

// StopHeartbeatTimer stops the heart beat timer (none leader scenario)
func (node *Node) StopHeartbeatTimer() {
	stopTimer(node.heartbeatTimer)
}

// ResetHeartbeatTimer resets the heart beat timer for the leader
func (node *Node) ResetHeartbeatTimer() {
	stopTimer(node.heartbeatTimer)
	node.heartbeatTimer.Reset(time.Duration(heartbeatTimeoutMs) * time.Millisecond)
}

// StartElection starts an election and elect self
func (node *Node) StartElection() bool {
	// set new candidate term
	node.term++

	// only start real election when we haven't voted for others for a higher term
	if node.lastVotedTerm < node.term {
		for i := range node.votes {
			node.votes[i] = false
		}
		// vote for self first
		node.votes[node.id] = true
		node.lastVotedTerm = node.term

		node.logger.Printf("\u270b T%d: Node%d starts election...\n", node.term, node.id)
		node.network.Broadcast(node.id, node.createRequestVoteMessage())
	} else {
		// We are a doomed candidate - alreayd voted for others for current term. Don't start election, wait for next term instead
		node.logger.Printf("\U0001F613 T%d: Node%d is a doomed candidate, waiting for next term...\n", node.term, node.id)
	}

	return true
}

// Vote for newer term and when we haven't voted for it yet
func (node *Node) Vote(electMsg *raftMessage) bool {
	if electMsg.term > node.term && electMsg.term > node.lastVotedTerm {
		node.lastVotedTerm = electMsg.term
		node.logger.Printf("\U0001f4e7 T%d: Node%d votes for Node%d \n", electMsg.term, node.id, electMsg.nodeID)
		node.network.Send(node.id, electMsg.nodeID, node.createVoteMessage(electMsg))
		return true
	}

	return false
}

// CountVotes counts votes received and decide whether we win
func (node *Node) CountVotes(ballotMsg *raftMessage) bool {
	if ballotMsg.data == node.id && ballotMsg.term == node.term {
		node.votes[ballotMsg.nodeID] = true

		totalVotes := 0
		for _, v := range node.votes {
			if v {
				totalVotes++
			}
		}

		if totalVotes > node.size/2 {
			// Won election, start heartbeat
			node.logger.Printf("\u2705 T%d: Node%d wins election\n", node.term, node.id)
			node.SendHeartbeat()
			return true
		}
	}

	return false
}

// AckHeartbeat acks a heartbeat message
func (node *Node) AckHeartbeat(hbMsg *raftMessage) bool {
	// handle heartbeat message with the same or newer term
	if hbMsg.term >= node.term {
		node.logger.Printf("\U0001f493 T%d: Node%d <- Node%d\n", hbMsg.term, node.id, hbMsg.nodeID)
		node.term = hbMsg.term
		return true
	}

	return false
}

// SendHeartbeat sends a heartbeat message
func (node *Node) SendHeartbeat() bool {
	node.network.Broadcast(node.id, node.createHeartBeatMessage())
	return true
}

// Start starts a node
func (node *Node) Start(wg *sync.WaitGroup) {
	node.logger.Printf("Node%d starting...\n", node.id)

	node.ResetElectionTimer()

	var msg *raftMessage
	quit := false
	msgCh, _ := node.network.GetRecvChannel(node.id)
	electCh := node.electionTimer.C
	hbCh := node.heartbeatTimer.C

	for !quit {
		select {
		case msg = <-msgCh:
		case <-electCh:
			msg = node.createStartElectionMessage()
		case <-hbCh:
			msg = node.createSendHeartBeatMessage()
		}

		// Do the real processing
		quit = node.processMessage(msg)
	}

	if wg != nil {
		wg.Done()
	}
}
