package raft

// MessageType type used by raft
type MessageType string

// allowed requestType values
const (
	MsgVote          = "Vote"
	MsgElect         = "Elect"
	MsgHeartbeat     = "Heartbeat"
	MsgStartElection = "StartElection" // dummy message to handle election timeout
	MsgSendHeartBeat = "SendHeartbeat" // dummy message to send heart beat
)

// Message object used by raft
type Message struct {
	nodeID  int
	term    int
	msgType MessageType
	data    int
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

func (node *Node) createStartElectionMessage() *Message {
	return &Message{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgStartElection,
		data:    node.id,
	}
}

func (node *Node) createHeartBeatMessage() *Message {
	return &Message{
		nodeID:  node.id,
		term:    node.term,
		msgType: MsgHeartbeat,
		data:    node.id,
	}
}
