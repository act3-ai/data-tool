// Package multiplex implements a multiplex and demultiplexer for streams of binary data
package multiplex

import (
	"bytes"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func TestMux(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	data1 := "abcdefghijklmnop"
	data2 := "0123456789end"

	bs := 4
	n := streamSize(len(data1), bs)
	n += streamSize(len(data2), bs)

	dst := &bytes.Buffer{}
	dst.Grow(n)

	src1 := bytes.NewBufferString(data1)
	src2 := bytes.NewBufferString(data2)

	assert.NoError(t, Mux(bs, dst, src1, src2))

	out := dst.Bytes()
	assert.Len(t, out, n)

	// try to Demux it
	d1 := &bytes.Buffer{}
	d1.Grow(len(data1))
	d2 := &bytes.Buffer{}
	d1.Grow(len(data2))
	assert.NoError(t, Demux(dst, d1, d2))

	assert.Equal(t, data1, d1.String())
	assert.Equal(t, data2, d2.String())
}

// streamSize computes the size of the multiplexed stream given data size (ds) and block size (bs).
func streamSize(ds, bs int) int {
	n := (headerSize + bs) * (ds / bs)
	if ds%bs != 0 {
		n += headerSize + ds%bs
	}
	return n
}

func TestDemux(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	src := bytes.NewBuffer([]byte{
		1, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 10, 11, 12, 13, 14, // chunk1
		0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 10, 11, // chunk2
		0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 10, 11, 12, 13, // chunk 3
	})

	d1 := &bytes.Buffer{}
	d2 := &bytes.Buffer{}
	assert.NoError(t, Demux(src, d1, d2))

	assert.Equal(t, []byte{10, 11, 10, 11, 12, 13}, d1.Bytes())
	assert.Equal(t, []byte{10, 11, 12, 13, 14}, d2.Bytes())
}

func BenchmarkMux(b *testing.B) {
	b.StopTimer()
	data1 := bytes.Repeat([]byte{1, 2, 3, 4}, 10000000)
	data2 := bytes.Repeat([]byte{5, 6, 7, 8}, 20000000)

	bs := 32 * 1024
	n := streamSize(len(data1), bs)
	n += streamSize(len(data2), bs)

	dst := &bytes.Buffer{}
	dst.Grow(n)

	for i := 0; i <= b.N; i++ {
		dst.Reset()
		src1 := bytes.NewBuffer(data1)
		src2 := bytes.NewBuffer(data2)

		b.StartTimer()
		err := Mux(bs, dst, src1, src2)
		b.StopTimer()

		assert.NoError(b, err)
		assert.Len(b, dst.Bytes(), n)
	}
	b.SetBytes(int64(len(data1)) + int64(len(data2)))
}

func BenchmarkDemux(b *testing.B) {
	b.StopTimer()
	data1 := bytes.Repeat([]byte{1, 2, 3, 4}, 10000000)
	data2 := bytes.Repeat([]byte{5, 6, 7, 8}, 20000000)

	bs := 32 * 1024
	n := streamSize(len(data1), bs)
	n += streamSize(len(data2), bs)

	dst := &bytes.Buffer{}
	dst.Grow(n)

	src1 := bytes.NewBuffer(data1)
	src2 := bytes.NewBuffer(data2)

	assert.NoError(b, Mux(bs, dst, src1, src2))

	data := dst.Bytes()

	d1 := &bytes.Buffer{}
	d1.Grow(len(data1))
	d2 := &bytes.Buffer{}
	d2.Grow(len(data2))

	for i := 0; i <= b.N; i++ {
		src := bytes.NewBuffer(data)
		d1.Reset()
		d2.Reset()

		b.StartTimer()
		err := Demux(src, d1, d2)
		b.StopTimer()

		assert.NoError(b, err)
		assert.Equal(b, data1, d1.Bytes())
		assert.Equal(b, data2, d2.Bytes())
	}
	b.SetBytes(int64(len(data1)) + int64(len(data2)))
}
