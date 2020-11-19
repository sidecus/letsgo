package raft

import (
	"math/rand"
	"time"
)

const minElectionTimeoutMs = 1500
const maxElectionTimeoutMs = 3000

type nodeState int

const (
	follower  = 0
	candidate = 1
	leader    = 2
)

// Node represents a raft node
type Node struct {
	cluster       *Cluster
	id            int
	leaderID      int
	term          int
	lastVotedTerm int
	votes         []int
	state         nodeState
	electionTimer *time.Timer
	recvChannel   chan *Message
}

// CreateNode creates a new node
func CreateNode(cluster *Cluster, id int, ch chan *Message) *Node {
	return &Node{
		cluster:       cluster,
		id:            id,
		leaderID:      0, // the initial value doesn't really matter
		term:          0,
		lastVotedTerm: 0,
		votes:         make([]int, cluster.size),
		state:         follower,
		recvChannel:   ch,
	}
}

func getTimeout() time.Duration {
	timeout := rand.Intn(maxElectionTimeoutMs-minElectionTimeoutMs) + minElectionTimeoutMs
	return time.Duration(timeout) * time.Millisecond
}

// resetTimer resets the election timer for the node
func (node *Node) resetTimer() {
	if node.electionTimer == nil {
		// create and start timer
		timer := time.NewTimer(getTimeout())
		node.electionTimer = timer
	} else {
		// stop, drain and reset timer
		if !node.electionTimer.Stop() {
			<-node.electionTimer.C
			node.electionTimer.Reset(getTimeout())
		}
	}
}

func (node *Node) createElectMessage() *Message {
	return &Message{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgElect,
		data:    node.id,
	}
}

func (node *Node) createVoteMessage(electMsg *Message) *Message {
	return &Message{
		nodeID:  node.id,
		term:    electMsg.term,
		msgType: MsgVote,
		data:    electMsg.nodeID,
	}
}

// move to candidate state and start election
func (node *Node) elect() {
	for i := range node.votes {
		node.votes[i] = 0
	}
	node.term++
	node.state = candidate
	node.lastVotedTerm = node.term
	node.cluster.network.Broadcast(node.id, node.createElectMessage())

	// Reset a new timer
	node.resetTimer()
}

func (node *Node) processMessage(msg *Message) {
	switch msg.msgType {
	case MsgVote:
	case MsgElect:
		node.processElectMessage(msg)
	case MsgHeartbeat:
	default:
		panic("Invalid message type")
	}
}

func (node *Node) processElectMessage(electMsg *Message) {
	// Only vote when we haven't voted for this term yet
	if node.lastVotedTerm < electMsg.term {
		node.lastVotedTerm = electMsg.term
		node.cluster.network.Send(node.id, electMsg.nodeID, node.createVoteMessage(electMsg))
	}
}

func (node *Node) processHeartBeatMessage(hbMsg *Message) {
	// Reset timer upon heart beat message
	if node.state == follower && hbMsg.term >= node.term {
		node.resetTimer()
	}
}

// Start starts a node
func (node *Node) Start() {
	node.resetTimer()

	go func() {
		select {
		case <-node.electionTimer.C:
			node.elect()
		case msg := <-node.recvChannel:
			node.processMessage(msg)
		}
	}()

	// whatever happens, new
	node.resetTimer()
}
