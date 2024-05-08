package archive

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/opencontainers/go-digest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeCompressibleData generates some random compressible data -- printable ascii characters in a somewhat
// random pattern with lots of repetition.
func makeCompressibleData(sz int) []byte {
	data := make([]byte, 0, sz)
	// I do not think we want this to be truly random since this is a test
	// s := rand.NewSource(rand.Int63())
	s := rand.NewSource(0)
	rng := rand.New(s)
	for i := 0; ; i++ {
		a := byte(rng.Intn(94) + 32)
		b := byte(rng.Intn(94) + 32)
		c := byte(rng.Intn(94) + 32)
		data = append(data, bytes.Repeat([]byte{a}, 40)...)
		data = append(data, bytes.Repeat([]byte{b}, 20)...)
		data = append(data, bytes.Repeat([]byte{c}, 40)...)
		if len(data) >= sz {
			break
		}
	}
	return data
}

// TestPipeCompressThreshold_Compressible tests that compressible data is properly continued after a threshold is
// reached for checking compression.  compressible data is generated, and compressed using the PipeCompressThreshold
// PipeWriter.  The compressed data is fed through a counter pipe writer, and the count is validated against a side
// band compressed version of the same input data, using size and digest.
func TestPipeCompressThreshold_Compressible(t *testing.T) {

	dataSz := 10000000
	data := makeCompressibleData(dataSz)
	odata := &BufWriteCloser{new(bytes.Buffer)}
	xdata := &BufWriteCloser{new(bytes.Buffer)}

	// create a compressor -- the PipeCompressThreshold object is used just to ensure the settings are consistent,
	// we use the compressor directly for this step
	pct := NewPipeCompressThreshold(3, 100000, false)
	writer := pct.createCompressor(xdata)

	// create a PipeCompressThreshold pipe writer to test threshold ratio checking and compression continuation.  The
	// data is funneled through a counter both to test proper writer forwarding and that the number of compressed bytes
	// meets expected values.
	pct = NewPipeCompressThreshold(3, 100000, false)
	counter := NewPipeCounter()
	counter.ConnectOut(odata)
	pct.ConnectOut(counter)

	// Since we're working with memory buffers for testing, writing the data directly will simply push the entire
	// buffer through the write routine at once.  Here we force a chunking of the data, more similar to what would
	// be encountered during an io operation to test the thresholding check during chunked writes
	rstart := 0
	rend := 320000
	for i := 0; ; i++ {
		if rend > dataSz {
			_, _ = pct.Write(data[rstart:])
			_, _ = writer.Write(data[rstart:])
			break
		}
		_, err := pct.Write(data[rstart:rend])
		assert.NoError(t, err)
		_, err = writer.Write(data[rstart:rend])
		assert.NoError(t, err)
		rstart = rend
		rend += 32000
	}

	err := pct.Close()
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	assert.Equalf(t, int(counter.Count), xdata.Len(), "Compressible data output length mismatch")
	assert.Equalf(t, digest.FromBytes(xdata.Bytes()), digest.FromBytes(odata.Bytes()), "Compressible digest mismatch")
}

// TestPipeCompressThreshold_CompressibleSmall tests that compressible data is properly forwarded after a threshold is
// reached for checking compression.  This specifically tests the case where the input data is smaller than both the
// read steps and check threshold, so the compression check is done on the close operation.  it is otherwise identical
// to TestPipeCompressThreshold_Compressible
func TestPipeCompressThreshold_CompressibleSmall(t *testing.T) {
	// Create small data for this test
	dataSz := 10000
	data := makeCompressibleData(dataSz)
	odata := &BufWriteCloser{new(bytes.Buffer)}
	xdata := &BufWriteCloser{new(bytes.Buffer)}

	// create a compressor -- the PipeCompressThreshold object is used just to ensure the settings are consistent,
	// we use the compressor directly for this step
	pct := NewPipeCompressThreshold(3, 100000, false)
	writer := pct.createCompressor(xdata)

	// create a PipeCompressThreshold pipe writer to test threshold ratio checking and compression continuation.  The
	// data is funneled through a counter both to test proper writer forwarding and that the number of compressed bytes
	// meets expected values.
	pct = NewPipeCompressThreshold(3, 100000, false)
	counter := NewPipeCounter()
	counter.ConnectOut(odata)
	pct.ConnectOut(counter)

	// Since we're working with memory buffers for testing, writing the data directly will simply push the entire
	// buffer through the write routine at once.  Here we force a chunking of the data, more similar to what would
	// be encountered during an io operation to test the thresholding check during chunked writes
	rstart := 0
	rend := 320000
	for i := 0; ; i++ {
		if rend > dataSz {
			_, _ = pct.Write(data[rstart:])
			_, _ = writer.Write(data[rstart:])
			break
		}
		_, err := pct.Write(data[rstart:rend])
		assert.NoError(t, err)
		_, err = writer.Write(data[rstart:rend])
		assert.NoError(t, err)
		rstart = rend
		rend += 32000
	}
	err := pct.Close()
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	assert.Equalf(t, int(counter.Count), xdata.Len(), "Compressible data output length mismatch")
	assert.Equalf(t, digest.FromBytes(xdata.Bytes()), digest.FromBytes(odata.Bytes()), "Compressible digest mismatch")

}

// TestPipeCompressThreshold_UnCompressible tests to ensure a poorly compressible data stream is forwarded through
// PipeCompressThreshold unchanged
func TestPipeCompressThreshold_UnCompressible(t *testing.T) {
	rng := rand.New(rand.NewSource(1))

	dataSz := 10000000
	data := make([]byte, dataSz)
	_, err := rng.Read(data)
	require.NoError(t, err)

	odata := &BufWriteCloser{new(bytes.Buffer)}
	xbuf := new(bytes.Buffer)

	// counter to ensure the output data count matches the expected count.  This also implicitly tests forwarding
	// output data to a second pipewriter
	counter := NewPipeCounter()
	counter.ConnectOut(odata)

	// Create a PipeCompressThreshold compressor with a small threshold for performing a check
	pct := NewPipeCompressThreshold(3, 100000, false)
	pct.ConnectOut(counter)

	// Since we're working with memory buffers for testing, writing the data directly will simply push the entire
	// buffer through the write routine at once.  Here we force a chunking of the data, more similar to what would
	// be encountered during an io operation to test the thresholding check during chunked writes
	rstart := 0
	rend := 320000
	for i := 0; ; i++ {
		if rend > dataSz {
			_, err := pct.Write(data[rstart:])
			assert.NoError(t, err)
			_, err = xbuf.Write(data[rstart:])
			assert.NoError(t, err)
			break
		}
		_, err := pct.Write(data[rstart:rend])
		assert.NoError(t, err)

		_, err = xbuf.Write(data[rstart:rend])
		assert.NoError(t, err)

		rstart = rend
		rend += 32000
	}
	err = pct.Close()
	assert.NoError(t, err)

	assert.Equalf(t, int(counter.Count), len(data), "Uncompressible data output length mismatch")
	assert.Equalf(t, digest.FromBytes(xbuf.Bytes()), digest.FromBytes(odata.Bytes()), "Uncompressible digest mismatch")
}

func TestPipeCompressThreshold_CompressContinue(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	data := make([]byte, 100000)
	_, err := rng.Read(data)
	require.NoError(t, err)

	odata := new(bytes.Buffer)
	odataA := new(bytes.Buffer)
	// odataB := new(bytes.Buffer)

	pct := NewPipeCompressThreshold(3, 50000, false)
	cmp := pct.createCompressor(odata)
	n, err := cmp.Write(data)
	if err != nil {
		t.Fatalf("err on write")
	}
	if n != len(data) {
		t.Fatalf("didn't write all data")
	}
	_ = cmp.Flush()

	partA := pct.createCompressor(odataA)
	// partB := pct.createCompressor(odataB)

	_, err = partA.Write(data[:50000])
	assert.NoError(t, err)

	err = partA.Flush()
	assert.NoError(t, err)

	_, err = partA.Write(data[50000:])
	assert.NoError(t, err)

	err = partA.Flush()
	assert.NoError(t, err)

	t.Log(n, odata.Len(), odataA.Len(), digest.FromBytes(odata.Bytes()), digest.FromBytes(odataA.Bytes()))
}

func TestPipeCompressThreshold_IsCompressible(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	makeBuffer := func(n int) BufWriteCloser {
		b := make([]byte, n)
		_, err := rng.Read(b)
		require.NoError(t, err)
		return BufWriteCloser{bytes.NewBuffer(b)}
	}

	type fields struct {
		uncBuf BufWriteCloser
		cmpBuf BufWriteCloser
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"sizes equal", fields{
			uncBuf: makeBuffer(10000),
			cmpBuf: makeBuffer(10000),
		}, false},
		{"compressed buffer larger", fields{
			uncBuf: makeBuffer(10000),
			cmpBuf: makeBuffer(11000),
		}, false},
		{"compressed buffer smaller but close", fields{
			uncBuf: makeBuffer(10000),
			cmpBuf: makeBuffer(99000),
		}, false},
		{"compressed buffer smaller", fields{
			uncBuf: makeBuffer(10000),
			cmpBuf: makeBuffer(9000),
		}, true},
		{"uncompressed buffer zero", fields{
			uncBuf: makeBuffer(0),
			cmpBuf: makeBuffer(9000),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct := &PipeCompressThreshold{
				uncBuf: tt.fields.uncBuf,
				cmpBuf: tt.fields.cmpBuf,
			}
			assert.Equalf(t, tt.want, pct.IsCompressible(), "IsCompressible()")
		})
	}
}
