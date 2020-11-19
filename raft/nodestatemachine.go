package raft

// TimerAction is the action we want to take on the given timer
type TimerAction int

const (
	timerNoop  = 0
	timerStop  = 1
	timerReset = 2
)

type msgHandler struct {
	handle               func(INode, *Message) bool
	nextState            NodeState
	electTimerAction     TimerAction
	heartbeatTimerAction TimerAction
}
type msgHandlerMap map[MessageType]msgHandler
type nodeStateMachine map[NodeState]msgHandlerMap

// processMessage runs a message through the node state machine
// if message is handled and state change required, it'll perform the needed work
// including state change and timer stop/reset
func (nodesm nodeStateMachine) processMessage(node INode, msg *Message) {
	handlerMap, validState := nodesm[node.State()]
	if !validState {
		panic("Invalid state for node %d")
	}

	handler, hasHandler := handlerMap[msg.msgType]
	if hasHandler && handler.handle != nil && handler.handle(node, msg) {
		// change state
		node.SetState(handler.nextState)

		// update election timer
		if handler.electTimerAction == timerStop {
			node.StopElectionTimer()
		} else if handler.electTimerAction == timerReset {
			node.ResetElectionTimer()
		}

		// update heartbreat timer
		if handler.heartbeatTimerAction == timerStop {
			node.StopHeartbeatTimer()
		} else if handler.heartbeatTimerAction == timerReset {
			node.ResetHeartbeatTimer()
		}
	}
}

func handleStartElection(node INode, msg *Message) bool {
	return node.Elect()
}

func handleSendHearbeat(node INode, msg *Message) bool {
	return node.SendHeartbeat()
}

func handleHeartbeat(node INode, msg *Message) bool {
	return node.AckHeartbeat(msg)
}

func handleBallotMsg(node INode, msg *Message) bool {
	return node.CountBallots(msg)
}

func handleElectMsg(node INode, msg *Message) bool {
	return node.Vote(msg)
}

// RaftNodeSM is the predefined node state machine
var RaftNodeSM = nodeStateMachine{
	follower: {
		MsgStartElection: {
			handle:               handleStartElection,
			nextState:            candidate,
			electTimerAction:     timerReset,
			heartbeatTimerAction: timerStop,
		},
		MsgHeartbeat: {
			handle:               handleHeartbeat,
			nextState:            follower,
			electTimerAction:     timerReset,
			heartbeatTimerAction: timerStop,
		},
		MsgElect: {
			handle:               handleElectMsg,
			nextState:            follower,
			electTimerAction:     timerNoop,
			heartbeatTimerAction: timerNoop,
		},
	},
	candidate: {
		MsgStartElection: {
			handle:               handleStartElection,
			nextState:            candidate,
			electTimerAction:     timerReset,
			heartbeatTimerAction: timerStop,
		},
		MsgHeartbeat: {
			handle:               handleHeartbeat,
			nextState:            follower,
			electTimerAction:     timerReset,
			heartbeatTimerAction: timerStop,
		},
		MsgElect: {
			handle:               handleElectMsg,
			nextState:            follower,
			electTimerAction:     timerNoop,
			heartbeatTimerAction: timerNoop,
		},
		MsgBallot: {
			handle:               handleBallotMsg,
			nextState:            leader,
			electTimerAction:     timerStop,
			heartbeatTimerAction: timerReset,
		},
	},
	leader: {
		MsgSendHeartBeat: {
			handle:               handleSendHearbeat,
			nextState:            leader,
			electTimerAction:     timerStop,
			heartbeatTimerAction: timerReset,
		},
		MsgHeartbeat: {
			handle:               handleHeartbeat,
			nextState:            follower,
			electTimerAction:     timerReset,
			heartbeatTimerAction: timerStop,
		},
	},
}
