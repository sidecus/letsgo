package raft

// timerAction is the action we want to take on the given timer
type timerAction int

const (
	timerActionNoop  = 0
	timerActionStop  = 1
	timerActionReset = 2
)

// raftMsgHandler defines a message handler struct
type raftMsgHandler struct {
	handle               func(INode, *raftMessage) bool
	nextState            NodeState
	electTimerAction     timerAction
	heartbeatTimerAction timerAction
}

//raftMsgHandlerMap defines map of message type to handler
type raftMsgHandlerMap map[raftMessageType]raftMsgHandler

// nodeStateMachine defines map from state to MsgHandlerMap
type nodeStateMachine map[NodeState]raftMsgHandlerMap

// ProcessMessage runs a message through the node state machine
// if message is handled and state change is required, it'll perform other needed work
// including advancing node state and stoping/reseting related timers
func (nodesm nodeStateMachine) ProcessMessage(node INode, msg *raftMessage) {
	handlerMap, validState := nodesm[node.State()]
	if !validState {
		panic("Invalid state for node %d")
	}

	handler, hasHandler := handlerMap[msg.msgType]
	if hasHandler && handler.handle != nil && handler.handle(node, msg) {
		// set new state
		node.SetState(handler.nextState)

		// update election timer
		if handler.electTimerAction == timerActionStop {
			node.StopElectionTimer()
		} else if handler.electTimerAction == timerActionReset {
			node.ResetElectionTimer()
		}

		// update heartbreat timer
		if handler.heartbeatTimerAction == timerActionStop {
			node.StopHeartbeatTimer()
		} else if handler.heartbeatTimerAction == timerActionReset {
			node.ResetHeartbeatTimer()
		}
	}
}

func handleStartElection(node INode, msg *raftMessage) bool {
	return node.StartElection()
}

func handleSendHearbeat(node INode, msg *raftMessage) bool {
	return node.SendHeartbeat()
}

func handleHeartbeat(node INode, msg *raftMessage) bool {
	return node.AckHeartbeat(msg)
}

func handleVoteMsg(node INode, msg *raftMessage) bool {
	return node.CountVotes(msg)
}

func handleRequestVoteMsg(node INode, msg *raftMessage) bool {
	return node.Vote(msg)
}

// raftNodeSM is the predefined node state machine
var raftNodeSM = nodeStateMachine{
	follower: {
		MsgStartElection: {
			handle:               handleStartElection,
			nextState:            candidate,
			electTimerAction:     timerActionReset,
			heartbeatTimerAction: timerActionStop,
		},
		MsgHeartbeat: {
			handle:               handleHeartbeat,
			nextState:            follower,
			electTimerAction:     timerActionReset,
			heartbeatTimerAction: timerActionStop,
		},
		MsgRequestVote: {
			handle:               handleRequestVoteMsg,
			nextState:            follower,
			electTimerAction:     timerActionNoop,
			heartbeatTimerAction: timerActionNoop,
		},
	},
	candidate: {
		MsgStartElection: {
			handle:               handleStartElection,
			nextState:            candidate,
			electTimerAction:     timerActionReset,
			heartbeatTimerAction: timerActionStop,
		},
		MsgHeartbeat: {
			handle:               handleHeartbeat,
			nextState:            follower,
			electTimerAction:     timerActionReset,
			heartbeatTimerAction: timerActionStop,
		},
		MsgRequestVote: {
			handle:               handleRequestVoteMsg,
			nextState:            follower,
			electTimerAction:     timerActionNoop,
			heartbeatTimerAction: timerActionNoop,
		},
		MsgVote: {
			handle:               handleVoteMsg,
			nextState:            leader,
			electTimerAction:     timerActionStop,
			heartbeatTimerAction: timerActionReset,
		},
	},
	leader: {
		MsgSendHeartBeat: {
			handle:               handleSendHearbeat,
			nextState:            leader,
			electTimerAction:     timerActionStop,
			heartbeatTimerAction: timerActionReset,
		},
		MsgHeartbeat: {
			handle:               handleHeartbeat,
			nextState:            follower,
			electTimerAction:     timerActionReset,
			heartbeatTimerAction: timerActionStop,
		},
		MsgRequestVote: {
			handle:               handleRequestVoteMsg,
			nextState:            follower,
			electTimerAction:     timerActionNoop,
			heartbeatTimerAction: timerActionNoop,
		},
	},
}
