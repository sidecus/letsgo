package raft

import (
	"math/rand"
	"time"
)

const minElectionTimeoutMs = 3500 // millisecond
const maxElectionTimeoutMs = 5000 // millisecond
const heartbeatTimeoutMs = 3600   // millisecond, larger value so that we can mimic some failures

// NodeState - node state type, can be one of 3: follower, candidate or leader
type NodeState int

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
	Elect() bool
	Vote(electMsg *Message) bool
	CountBallots(ballotMsg *Message) bool
	AckHeartbeat(hbMsg *Message) bool
	SendHeartbeat() bool
	Start()
}

// Node represents a raft node
type Node struct {
	cluster        *Cluster
	id             int
	term           int
	lastVotedTerm  int
	votes          []bool
	state          NodeState
	electionTimer  *time.Timer // timer for election timeout, used by follower and candidate
	heartbeatTimer *time.Timer // timer for heartbeat, used by leader
	recvChannel    chan *Message
	sm             NodeStateMachine
}

// CreateNode creates a new node
func CreateNode(cluster *Cluster, id int, ch chan *Message) *Node {
	// Initialize timer objects (stopped immediately)
	electionTimer := time.NewTimer(time.Hour)
	electionTimer.Stop()
	heartbeatTimer := time.NewTimer(time.Hour)
	heartbeatTimer.Stop()

	return &Node{
		cluster:        cluster,
		id:             id,
		term:           0,
		lastVotedTerm:  0,
		votes:          make([]bool, cluster.size),
		state:          follower,
		electionTimer:  electionTimer,
		heartbeatTimer: heartbeatTimer,
		recvChannel:    ch,
		sm:             RaftNodeSM,
	}
}

func getElectionTimeout() time.Duration {
	timeout := rand.Intn(maxElectionTimeoutMs-minElectionTimeoutMs) + minElectionTimeoutMs
	return time.Duration(timeout) * time.Millisecond
}

// stop and drain the timer
func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		// The Timer document is inaccurate with a bad exmaple - timer.Stop returning false doesn't necessarily
		// mean there is anything to drain in the channel. Blind draining can cause dead locking
		// e.g. Stop is called after event is already fired. In this case draining the channel will block.
		// We use a default clause here to stop us from blocking - the current go routine is the sole reader of the timer channel
		// so no synchronization is required
		select {
		case <-timer.C:
		default:
		}
	}
}

// sends a heartbreat message
func (node *Node) heartbeat() {
	node.cluster.network.Broadcast(node.id, node.createHeartBeatMessage())
}

// processMessage passes the message through the node statemachine
func (node *Node) processMessage(msg *Message) {
	node.sm.ProcessMessage(node, msg)
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
	node.electionTimer.Reset(getElectionTimeout())
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

// Elect starts an election and elect self
func (node *Node) Elect() bool {
	for i := range node.votes {
		node.votes[i] = false
	}
	// vote for self directly
	node.votes[node.id] = true
	node.term++
	node.lastVotedTerm = node.term
	node.cluster.logger.Printf("Node%d starts T%d election...\n", node.id, node.term)
	node.cluster.network.Broadcast(node.id, node.createElectMessage())

	return true
}

// Vote for newer term and when we haven't voted for it yet
func (node *Node) Vote(electMsg *Message) bool {
	if electMsg.term > node.term && electMsg.term > node.lastVotedTerm {
		node.lastVotedTerm = electMsg.term
		node.cluster.logger.Printf("Node%d votes on T%d for Node%d \n", node.id, electMsg.term, electMsg.nodeID)
		node.cluster.network.Send(node.id, electMsg.nodeID, node.createBallotMessage(electMsg))
		return true
	}

	return false
}

// CountBallots counts ballots received and decide whether we win
func (node *Node) CountBallots(ballotMsg *Message) bool {
	if ballotMsg.data == node.id && ballotMsg.term == node.term {
		node.votes[ballotMsg.nodeID] = true

		totalVotes := 0
		for _, v := range node.votes {
			if v {
				totalVotes++
			}
		}

		if totalVotes > node.cluster.size/2 {
			// Won election, start heartbeat
			node.cluster.logger.Printf("Node%d wins T%d election (%d votes)...\n", node.id, node.term, totalVotes)
			node.heartbeat()
			return true
		}
	}

	return false
}

// AckHeartbeat acks a heartbeat message
func (node *Node) AckHeartbeat(hbMsg *Message) bool {
	// handle heartbeat message with the same or newer term
	if hbMsg.term >= node.term {
		node.cluster.logger.Printf("Node%d acks T%d heartbeat from Node%d\n", node.id, hbMsg.term, hbMsg.nodeID)
		node.term = hbMsg.term
		return true
	}

	return false
}

// SendHeartbeat sends a heartbeat message
func (node *Node) SendHeartbeat() bool {
	node.heartbeat()
	return true
}

// Start starts a node
func (node *Node) Start() {
	node.cluster.logger.Printf("Node%d starting...\n", node.id)

	go func() {
		node.ResetElectionTimer()
		var msg *Message
		for {
			select {
			case <-node.electionTimer.C:
				// We should trigger election
				msg = node.createStartElectionMessage()
			case <-node.heartbeatTimer.C:
				// We should trigger heartbeat
				msg = node.createSendHeartBeatMessage()
			case msg = <-node.recvChannel:
				// We should handle the incoming message
			}

			// Do the real processing
			node.processMessage(msg)
		}
	}()
}
