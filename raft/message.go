package raft

// raftMessageType message type used by raft nodes
type raftMessageType string

// allowed raftMessageType values
const (
	MsgVote          = "Vote"
	MsgRequestVote   = "RequestVote"
	MsgHeartbeat     = "Heartbeat"
	MsgStartElection = "StartElection" // dummy message to handle new election
	MsgSendHeartBeat = "SendHeartbeat" // dummy message to send heartbeat
)

// raftMessage object used by raft
type raftMessage struct {
	nodeID  int
	term    int
	msgType raftMessageType
	data    int
}

func (node *Node) createRequestVoteMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgRequestVote,
		data:    node.id,
	}
}

func (node *Node) createStartElectionMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgStartElection,
		data:    node.id,
	}
}

func (node *Node) createVoteMessage(electMsg *raftMessage) *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    electMsg.term,
		msgType: MsgVote,
		data:    electMsg.nodeID,
	}
}

func (node *Node) createHeartBeatMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgHeartbeat,
		data:    node.id,
	}
}

func (node *Node) createSendHeartBeatMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgSendHeartBeat,
		data:    node.id,
	}
}
