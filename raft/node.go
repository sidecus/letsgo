package raft

import (
	"math/rand"
	"time"
)

const minElectionTimeoutMs = 5000
const maxElectionTimeoutMs = 6000

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
	term          int
	lastVotedTerm int
	votes         []bool
	state         nodeState
	electionTimer *time.Timer
	recvChannel   chan *Message
}

type msgHandler struct {
	handle    func(*Node, *Message) bool
	nextState nodeState
}
type msgHandlerMap map[MessageType]msgHandler
type nodeStateHandlers map[nodeState]msgHandlerMap

// CreateNode creates a new node
func CreateNode(cluster *Cluster, id int, ch chan *Message) *Node {
	return &Node{
		cluster:       cluster,
		id:            id,
		term:          0,
		lastVotedTerm: 0,
		votes:         make([]bool, cluster.size),
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
		node.electionTimer = time.NewTimer(getTimeout())
	} else {
		// stop, drain and reset timer
		if !node.electionTimer.Stop() {
			// We have to use select here with the default clause
			// The Timer document is incorrect, Stop returning false doesn't means there is anything in the channel,
			// e.g. when Stop is called after event is already fired
			// The default clause will stop us from blocking on it under that situation
			select {
			case <-node.electionTimer.C:
			default:
			}
		}
		node.electionTimer.Reset(getTimeout())
	}
}

// start election
func elect(node *Node, electMsg *Message) bool {
	for i := range node.votes {
		node.votes[i] = false
	}
	node.votes[node.id] = true
	node.term++
	node.lastVotedTerm = node.term
	node.cluster.logger.Printf("Node%d starts T%d election...\n", node.id, node.term)
	node.cluster.network.Broadcast(node.id, node.createElectMessage())

	// Reset a new timer
	node.resetTimer()

	return true
}

func vote(node *Node, electMsg *Message) bool {
	// Only vote for newer term and when we haven't voted for it yet
	if node.lastVotedTerm < electMsg.term && node.term < electMsg.term {
		node.lastVotedTerm = electMsg.term
		node.cluster.logger.Printf("Node%d votes on T%d for Node%d \n", node.id, electMsg.term, electMsg.nodeID)
		node.cluster.network.Send(node.id, electMsg.nodeID, node.createVoteMessage(electMsg))
		return true
	}

	return false
}

func countVote(node *Node, electMsg *Message) bool {
	if electMsg.data == node.id && electMsg.term == node.term {
		node.votes[electMsg.nodeID] = true

		totalVotes := 0
		for _, v := range node.votes {
			if v {
				totalVotes++
			}
		}

		if totalVotes >= node.cluster.size/2+1 {
			// Elected, send heart beat
			node.cluster.logger.Printf("Node%d wins T%d election (%d votes)...\n", node.id, node.term, totalVotes)
			node.cluster.network.Broadcast(node.id, node.createHeartBeatMessage())
			return true
		}
	}

	return false
}

func ackHeartbeat(node *Node, hbMsg *Message) bool {
	// Reset timer upon heart beat message
	if hbMsg.term >= node.term {
		node.cluster.logger.Printf("Node%d acks T%d heartbeat from Node%d\n", node.id, hbMsg.term, hbMsg.nodeID)
		node.term = hbMsg.term
		node.resetTimer()
		return true
	}

	return false
}

var nodeStateMachine = nodeStateHandlers{
	follower: {
		MsgStartElection: {handle: elect, nextState: candidate},
		MsgHeartbeat:     {handle: ackHeartbeat, nextState: follower},
		MsgElect:         {handle: vote, nextState: follower},
	},
	candidate: {
		MsgStartElection: {handle: elect, nextState: candidate},
		MsgElect:         {handle: vote, nextState: follower},
		MsgVote:          {handle: countVote, nextState: leader},
		MsgHeartbeat:     {handle: ackHeartbeat, nextState: follower},
	},
	leader: {
		MsgSendHeartBeat: {handle: nil, nextState: leader},
		MsgHeartbeat:     {handle: ackHeartbeat, nextState: follower},
	},
}

func (node *Node) processMessage(msg *Message) {
	handlerMap, validState := nodeStateMachine[node.state]
	if !validState {
		panic("Invalid state for node %d")
	}

	handler, hasHandler := handlerMap[msg.msgType]
	if hasHandler && handler.handle != nil && handler.handle(node, msg) {
		node.state = handler.nextState
	}
}

// Start starts a node
func (node *Node) Start() {
	node.cluster.logger.Printf("Node%d starting...\n", node.id)

	go func() {
		node.resetTimer()
		var msg *Message
		for {
			select {
			case <-node.electionTimer.C:
				msg = node.createStartElectionMessage()
			case msg = <-node.recvChannel:
			}
			node.processMessage(msg)
		}
	}()
}
