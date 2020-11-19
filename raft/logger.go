package raft

// Logger defines the raft logger interface
type Logger interface {
	Log(string)
}
