package raft

/* raft request type definition */

// MessageType type used by raft
type MessageType string

// allowed requestType values
const (
	MsgVote      = "V"
	MsgElect     = "E"
	MsgHeartbeat = "H"
)

// Message object used by raft
type Message struct {
	nodeID  int
	term    int
	msgType MessageType
	data    int // dummy data
}
