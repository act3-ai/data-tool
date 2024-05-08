package util

import "context"

// UnlimitedChannel creates an input/output channel pair that provides the behavior of an unlimited buffered channel.
// the internal go routine blocks on waiting for items to appear on the input channel, and adds them to a list of
// items when one is received.
func UnlimitedChannel[T any](cancel context.Context) (in chan T, out chan T) {
	in = make(chan T)
	out = make(chan T)
	go func() {
		var itemQueue []T
		// pop returns the item on the top of the queue, or nil if the queue is empty, preventing an attempt to access
		// the empty queue when trying to send items from an empty queue.
		pop := func() T {
			if len(itemQueue) == 0 {
				// Return the zero value for T
				var result T
				return result
			}
			return itemQueue[0]
		}
		// getOut is a local function that returns the output channel if items are in the queue, or nil otherwise.
		// this allows output to wait if nothing's on the queue versus writing nil values to the channel
		getOut := func() chan T {
			if len(itemQueue) == 0 {
				return nil
			}
			return out
		}
		// continue looping until the queue is drained or the input channel remains open
		for len(itemQueue) > 0 || in != nil {
			// this inner select blocks when nothing is on the input channel AND nothing is reading from the output
			// channel.
			select {
			// first check if an item waiting on the input channel
			case it, ok := <-in:
				// if ok is false, the input channel has been closed, and we should quit
				if !ok {
					in = nil
				} else {
					// item found, drain the input channel
					itemQueue = append(itemQueue, it)
				}
			// second, try writing an item to the output channel.  The local functions avoid reading from an empty
			// queue with pop, and writing nil values to a channel with getOut
			case getOut() <- pop():
				itemQueue = itemQueue[1:]
			case <-cancel.Done():
				in = nil
				return
			}
		}
		close(out)
	}()
	return
}
