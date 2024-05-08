// Package blockbuf implements a functionality similar to "mbuffer".
// This is useful for writing to tape drives that require large block sizes and continuous writes (without breaks).
package blockbuf

import (
	"errors"
	"io"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Copy will copy from src to dst but ensures that writes are always of block size (bs) except for possibly the last write.
// Writes are not attempted until the high water mark (hwm) number of blocks are first read.
// At which point the buffer is written until empty.  Then it waits until the high water mark is achieved again.
// The maximum buffer size is `n` blocks.
func Copy(dst io.Writer, src io.Reader, n, bs, hwm int) error {
	// cap the high water mark
	if n < hwm {
		hwm = n
	}

	c := newCopier(dst, src, n, bs, hwm)

	g := errgroup.Group{}
	g.Go(c.writer)
	g.Go(c.reader)

	if err := g.Wait(); err != nil {
		return err
	}

	// write all remaining blocks
	for block := range c.blocks {
		if err := c.writeBlock(block); err != nil {
			return err
		}
	}

	return nil
}

type copier struct {
	dst          io.Writer
	src          io.Reader
	n, bs, hwm   int
	blocks       chan []byte
	startWriting chan struct{}
	bufPool      *sync.Pool
}

func newCopier(dst io.Writer, src io.Reader, n, bs, hwm int) *copier {
	return &copier{
		dst: dst,
		src: src,
		n:   n,
		bs:  bs,
		hwm: hwm,

		blocks: make(chan []byte, n),

		// This channel is used to tell the writer it must start its work if not already doing so.
		// We make the channel large enough so that sending will never block.
		startWriting: make(chan struct{}, n-hwm),

		bufPool: &sync.Pool{
			New: func() any {
				return make([]byte, bs)
			},
		},
	}
}

func (c *copier) writeBlock(block []byte) error {
	defer c.bufPool.Put(block) //nolint:staticcheck
	if _, err := c.dst.Write(block); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func (c *copier) writer() error {
	for {
		// wait for the writer to let us start writing
		_, ok := <-c.startWriting
		if !ok {
			// the channel is closed
			return nil
		}

		// we can receive more "tokens" than expected so we double check the length of blocks.
		if len(c.blocks) < c.hwm {
			continue
		}

		// we got a token so start writing until the buffer is empty

		// non-blocking for-loop
	loop:
		for {
			select {
			case block, ok := <-c.blocks:
				if !ok {
					// channel is closed
					return nil
				}
				// we have a block to write
				if err := c.writeBlock(block); err != nil {
					return err
				}
			default:
				// channel is empty but not closed
				// more blocks might be available later
				break loop
			}
		}
	}
}

func (c *copier) reader() error {
	for {
		// read a full block from src
		// block := make([]byte, bs)
		block := c.bufPool.Get().([]byte)[:c.bs]
		if len(block) != c.bs {
			panic("Block of unexpected size received from pool")
		}

		if nr, er := io.ReadFull(c.src, block); er != nil {
			if errors.Is(er, io.EOF) {
				break
			}
			if errors.Is(er, io.ErrUnexpectedEOF) {
				// partial block, the last block since EOF was received on the source
				c.blocks <- block[:nr]
				break
			}
			return er //nolint:wrapcheck
		}
		// full size block

		// put it on the blocks channel
		c.blocks <- block

		// check if the high watermark was reached
		if len(c.blocks) >= c.hwm {
			// allow the writer to write
			// must ensure the writer received the signal otherwise we can deadlock on a full blocks channel
			c.startWriting <- struct{}{}
		}
	}

	// cleanup by making sure we write if needed
	close(c.blocks)
	close(c.startWriting) // should kill the writer after it is done with its session of writing.

	return nil
}
