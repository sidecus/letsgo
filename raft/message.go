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

func (node *raftNode) createRequestVoteMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgRequestVote,
		data:    node.id,
	}
}

func (node *raftNode) createStartElectionMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgStartElection,
		data:    node.id,
	}
}

func (node *raftNode) createVoteMessage(electMsg *raftMessage) *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    electMsg.term,
		msgType: MsgVote,
		data:    electMsg.nodeID,
	}
}

func (node *raftNode) createHeartBeatMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgHeartbeat,
		data:    node.id,
	}
}

func (node *raftNode) createSendHeartBeatMessage() *raftMessage {
	return &raftMessage{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgSendHeartBeat,
		data:    node.id,
	}
}
