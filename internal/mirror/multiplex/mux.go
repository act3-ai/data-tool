// Package multiplex implements a multiplex and demultiplexer for streams of binary data
package multiplex

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"golang.org/x/sync/errgroup"
)

/*
This code prevents head of line blocking (HOL) from slow readers but multiplexing them over a single writer.

From the shell this can be used like so:

multiplexer <(serialize reg1.example.com) <(serialize reg2.example.com) --block-size=1Mi | demultiplexer >(deserialize reg1) >(deserialize reg2)
*/

const headerSize = 4 + 8

// Mux will multiplex all the readers onto the same output writer.
// This can be demultiplexed with Demux.
func Mux(bs int, w io.Writer, readers ...io.Reader) error {
	g := errgroup.Group{}

	chunks := make(chan []byte) // unbuffered

	bufPool := &sync.Pool{
		New: func() any {
			return make([]byte, headerSize+bs)
		},
	}

	// reader group
	rg := errgroup.Group{}
	// kick off the reader goroutines
	for i, r := range readers {
		rg.Go(func() error {
			return read(i, r, chunks, bufPool)
		})
	}

	// write goroutine
	g.Go(func() error {
		return write(chunks, w, bufPool)
	})

	// goroutine to wait on the readers to finish then close the channel
	// This will allow the writer goroutine to finish
	g.Go(func() error {
		defer close(chunks)
		return rg.Wait()
	})

	return g.Wait()
}

func read(id int, src io.Reader, out chan<- []byte, bufPool *sync.Pool) error {
	for i := 0; ; i++ {
		// get a buffer from the pool
		buf := bufPool.Get().([]byte)

		// grow the slice back to capacity (which is always headerSize + bs)
		buf = buf[:cap(buf)]

		// we read directly into the buffer to avoid a copy
		nr, er := src.Read(buf[headerSize:])
		if nr > 0 {
			// fill in the header directly into the buffer to avoid a copy
			binary.LittleEndian.PutUint32(buf[:4], uint32(id))
			binary.LittleEndian.PutUint64(buf[4:], uint64(nr))
			out <- buf[:headerSize+nr]
		}
		if er != nil {
			if er != io.EOF {
				// non EOF error
				return er //nolint:wrapcheck
			}
			break
		}
	}
	return nil
}

func write(in <-chan []byte, w io.Writer, bufPool *sync.Pool) error {
	for buf := range in {
		if _, err := w.Write(buf); err != nil {
			return fmt.Errorf("writing data: %w", err)
		}

		// release the buffer back to the pool
		bufPool.Put(buf) //nolint:staticcheck
	}
	return nil
}
