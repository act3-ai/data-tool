package util

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestUnlimitedChannel tests an unbounded buffered channel, this version just ensures the output is complete and in
// the right order, with no delays to check asynchronous behavior.
func Test_UnlimitedChannel(t *testing.T) {
	in, out := UnlimitedChannel[int](context.Background())
	var wg sync.WaitGroup
	lastVal := -1
	wg.Add(1)
	go func() {
		for i := range out {
			if i != lastVal+1 {
				t.Errorf("Incorrect output order: got %d, want %d", i, lastVal+1)
				wg.Done()
			}
			lastVal = i
		}
		wg.Done()
	}()

	for i := 0; i < 100; i++ {
		in <- i
	}
	close(in)
	wg.Wait()
	if lastVal != 99 {
		t.Errorf("Did not get all values, stopped at %d", lastVal)
	}
}

// TestUnlimitedChannelNoReadBlock tests that reading from the output channel while empty can be non-blocking using
// a select statement. Note, this is standard go syntax, but this test remains as a reminder for read considerations.
func Test_UnlimitedChannelNoReadBlock(t *testing.T) {
	in, out := UnlimitedChannel[int](context.Background())
	select {
	case v := <-out:
		t.Errorf("Got non nil value from empty chan %v", v)
	default:
	}
	close(in)
}

// TestUnlimitedChannelReadBlock tests to ensure that reads from the output channel block until something arrives on
// input.  this is standard go behavior, but this test remains as a consideration for implementation when blocking is
// desired.
func Test_UnlimitedChannelReadBlock(t *testing.T) {
	in, out := UnlimitedChannel[int](context.Background())
	outval := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		v := <-out
		if v == 0 {
			t.Errorf("channel read did not block, nil was returned")
		}
		outval = v
		wg.Done()
	}()
	// sleep for a bit before adding anything on the input, giving the read routine a moment to try a read and block
	time.Sleep(200 * time.Millisecond)
	in <- 10
	wg.Wait()
	if outval != 10 {
		t.Errorf("output did not arrive at end of channel")
	}
	close(in)
}

// TestUnlimitedChannelCloseUnblock tests to ensure that reads from the output are unblocked when the channel is closed
// this is standard go behavior, but this test is left for consideration during implementation
func Test_UnlimitedChannelCloseUnblock(t *testing.T) {
	in, out := UnlimitedChannel[int](context.Background())
	var v int
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		v = <-out
		if v != 0 {
			t.Errorf("channel read returned unexpected value %v", v)
		}
		wg.Done()
	}()
	time.Sleep(200 * time.Millisecond)
	close(in)
	wg.Wait()
	if v != 0 {
		t.Errorf("channel read returned unexpected value %v", v)
	}
}

// TestUnlimitedChannelAsyncWrite tests the case where writes are slower than reads -- eg, the channel is drained as
// soon as it's filled, checking for empty queue error edge cases
func Test_UnlimitedChannelAsyncWrite(t *testing.T) {
	in, out := UnlimitedChannel[int](context.Background())
	var wg sync.WaitGroup
	lastVal := -1
	wg.Add(1)
	go func() {
		for i := range out {
			if i != lastVal+1 {
				t.Errorf("Incorrect output order: got %d, want %d", i, lastVal+1)
				wg.Done()
			}
			lastVal = i
		}
		wg.Done()
	}()

	// fmt.Println("Write Started")
	for i := 0; i < 20; i++ {
		in <- i
		time.Sleep(50 * time.Millisecond)
	}
	close(in)
	// fmt.Println("Done Writing")
	wg.Wait()
	if lastVal != 19 {
		t.Errorf("Did not get all values, stopped at %d", lastVal)
	}
}

// TestUnlimitedChannelAsyncRead tests cases where reads are much slower than writes, ensuring that writes are unblocked
// while the queue fills and that it is eventually drained even if the input channel is closed
func Test_UnlimitedChannelAsyncRead(t *testing.T) {
	in, out := UnlimitedChannel[int](context.Background())
	var wg sync.WaitGroup
	lastVal := -1
	wg.Add(1)
	go func() {
		for i := range out {
			if i != lastVal+1 {
				t.Errorf("Incorrect output order: got %d, want %d", i, lastVal+1)
				wg.Done()
			}
			lastVal = i
			time.Sleep(50 * time.Millisecond)
		}
		wg.Done()
	}()

	// fmt.Println("Write Started")
	for i := 0; i < 20; i++ {
		in <- i
	}
	close(in)
	// fmt.Println("Done Writing")
	wg.Wait()
	if lastVal != 19 {
		t.Errorf("Did not get all values, stopped at %d", lastVal)
	}
}
