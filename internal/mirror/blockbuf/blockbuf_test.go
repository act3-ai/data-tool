package blockbuf

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	data := "abcdefghijklmnopqr"

	dst := &bytes.Buffer{}
	dst.Grow(len(data))

	src := bytes.NewBufferString(data)

	assert.NoError(t, Copy(dst, src, 3, 4, 2))

	assert.Equal(t, data, dst.String())
}

func TestCopy_controlled(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	testCases := []struct {
		name string // name is a human-readable description of the test case, used to identify the test when reporting results.
		data string // data is the input data that will be written to the src buffer, simulating the data source.
		n    int    // n is the capacity of the blocks channel, which determines the maximum number of buffered blocks.
		bs   int    // bs is the block size, defining the size of each block being written from src to dst.
		hwm  int    // hwm is the high water mark, indicating the number of blocks that must be read before writing is triggered.
	}{
		{
			name: "check properly sized writes",
			data: "abcdefghijklmnopqr",
			n:    3,
			bs:   4,
			hwm:  2,
		},
		{
			name: "empty input",
			data: "",
			n:    3,
			bs:   4,
			hwm:  2,
		},
		{
			name: "single block",
			data: "abcdef",
			n:    3,
			bs:   6,
			hwm:  2,
		},
		{
			name: "partial block",
			data: "abcdefghijklmnop",
			n:    3,
			bs:   5,
			hwm:  2,
		},
		{
			name: "multiple full blocks",
			data: "abcdefghijklmnopqrst",
			n:    3,
			bs:   4,
			hwm:  2,
		},
		{
			name: "empty data",
			data: "",
			n:    3,
			bs:   4,
			hwm:  2,
		},
		{
			name: "block size equals data length",
			data: "abcdefgh",
			n:    3,
			bs:   8,
			hwm:  2,
		},
		{
			name: "hwm equals n",
			data: "abcdefghijkl",
			n:    3,
			bs:   4,
			hwm:  3,
		},
		{
			name: "hwm larger than n",
			data: "abcdefghijklmnop",
			n:    2,
			bs:   4,
			hwm:  3,
		},
		{
			name: "data length is a multiple of block size",
			data: "abcdefghijk",
			n:    3,
			bs:   4,
			hwm:  2,
		},
		{
			name: "hwm == n",
			data: "abcdefghijk",
			n:    3,
			bs:   1,
			hwm:  3,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// for this test we need to intercept all dst.Write() calls and make sure the writes are properly sized by the block size
			dst := &bytes.Buffer{}
			dst.Grow(len(tc.data))

			src := bytes.NewBufferString(tc.data)

			lastBlockSize := len(tc.data) % tc.bs
			if lastBlockSize == 0 {
				lastBlockSize = tc.bs
			}

			totalBlocks := (len(tc.data) + tc.bs - 1) / tc.bs

			cw := &checkedWriter{t: t, w: dst, expectedSize: tc.bs, lastBlockSize: lastBlockSize, totalBlocks: totalBlocks}
			assert.NoError(t, Copy(cw, src, tc.n, tc.bs, tc.hwm))

			assert.Equal(t, tc.data, dst.String())
		})
	}
}

type checkedWriter struct {
	t             *testing.T
	w             io.Writer
	i             int // i is a counter used to track the current block index being written, helping determine the expected block size.
	totalBlocks   int // totalBlocks is the total number of blocks that will be written.
	event         chan int
	expectedSize  int // expectedSize is the expected size of the current block being written.
	lastBlockSize int // lastBlockSize is the size of the last block written, used to verify that the last block is properly sized.
}

func (cw *checkedWriter) Write(p []byte) (n int, err error) {
	cw.t.Helper()
	switch {
	case cw.i < cw.totalBlocks-1:
		// Case 1: i < expected number of blocks, checking the block matches the expected block size.
		assert.Len(cw.t, p, cw.expectedSize)
	case cw.i == cw.totalBlocks-1:
		// Case 2: i == expected last block, checking the size of the last block like a "remainder".
		assert.Len(cw.t, p, cw.lastBlockSize)
	default:
		// Case 3: i > last block, asserting the fail of extra block written.
		assert.Fail(cw.t, "extra block written", string(p))
	}
	cw.i++

	n, err = cw.w.Write(p)
	if cw.event != nil {
		cw.event <- len(p)
	}
	return
}

func TestCopy_hwm(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	// for this test we are trying to check that the high water mark is working
	data := "abcdefghijklmnopqr"

	dst := &bytes.Buffer{}
	dst.Grow(len(data))
	cw := &checkedWriter{
		t:             t,
		w:             dst,
		totalBlocks:   5,
		expectedSize:  4,
		lastBlockSize: 2,
		event:         make(chan int),
	}
	src, srcWriter := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t.Helper()
		assert.NoError(t, Copy(cw, src, 3, 4, 2))
		wg.Done()
	}()

	_, err := srcWriter.Write([]byte("ab"))
	assert.NoError(t, err)
	// no write to dst
	assert.Equal(t, "", dst.String())

	_, err = srcWriter.Write([]byte("cd"))
	assert.NoError(t, err)
	// no write to dst
	assert.Equal(t, "", dst.String())

	_, err = srcWriter.Write([]byte("efg"))
	assert.NoError(t, err)
	// no write to dst
	assert.Equal(t, "", dst.String())

	_, err = srcWriter.Write([]byte("hij"))
	assert.NoError(t, err)
	// wait for dst.Write() to be called with two blocks
	assert.Equal(t, 4, <-cw.event)
	assert.Equal(t, 4, <-cw.event)
	assert.Equal(t, "abcdefgh", dst.String())

	time.Sleep(100 * time.Microsecond) // need to sleep to ensure that the following write does not happen before the write go routine is completed
	_, err = srcWriter.Write([]byte("klm"))
	assert.NoError(t, err)
	// no write to dst
	assert.Equal(t, "abcdefgh", dst.String())

	_, err = srcWriter.Write([]byte("nopqr"))
	assert.NoError(t, err)
	// wait for dst.Write() to be called with two blocks
	assert.Equal(t, 4, <-cw.event)
	assert.Equal(t, 4, <-cw.event)
	assert.Equal(t, "abcdefghijklmnop", dst.String())

	err = srcWriter.Close()
	assert.NoError(t, err)
	assert.Equal(t, 2, <-cw.event)
	wg.Wait()
	close(cw.event)
	assert.Equal(t, data, dst.String())
}

func TestCopy_big(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	data := bytes.Repeat([]byte("abcdefghijklmnopqr"), 10000)

	dst := &bytes.Buffer{}
	dst.Grow(len(data))

	src := bytes.NewBuffer(data)

	assert.NoError(t, Copy(dst, src, 100, 1024*1024, 90))

	assert.Equal(t, data, dst.Bytes())
}

func BenchmarkCopy(b *testing.B) {
	b.StopTimer()
	data := bytes.Repeat([]byte{1, 2, 3, 4}, 10000000)
	dst := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		dst.Reset()
		src := bytes.NewBuffer(data)

		b.StartTimer()
		err := Copy(dst, src, 10, 1024*1024, 9)
		b.StopTimer()

		assert.NoError(b, err)
		assert.Equal(b, data, dst.Bytes())
	}
	b.SetBytes(int64(len(data)))
}
