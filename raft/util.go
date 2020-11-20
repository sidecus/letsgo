package raft

import "time"

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
