package archive

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
)

// compressionRatioThreshold is a compression ratio (uncompressed size / compressed size) that will be used to determine if compression should be applied.
// If we do not save at least 10% then compression is probably not worth compressing it (because the receiver has to decompress it).
const compressionRatioThreshold = 1.10

// PipeWriter is an interface that provides the necessary functionality for operating with output pipe streams.  This
// provides a io.WriteCloser compatible interface, and thus can be used as the destination for any stream write
// operation.  In addition, the output can be connected to other write closers (commonly, other PipeWriter
// implementations), to daisy chain a collection of write handlers.
type PipeWriter interface {
	io.Writer
	io.Closer
	// ConnectOut wraps a WriteCloser with additional functionality that can occur during write or close with the
	// expectation that the data is forwarded down stream to the wrapped writer, as well as close operations
	ConnectOut(w io.WriteCloser) PipeWriter
}

// PipeReader is an interface that provides the necessary functionality for operating with input pipe streams.  This
// provides a io.ReadCloser compatible interface, and thus can be used as the source for any stream read operation.  In
// addition, the input can be connected to other read closers (commonly, other PipeReader implementations), to daisy
// chain a collection of read handlers.
type PipeReader interface {
	io.Reader
	io.Closer
	// ConnectIn wraps a ReadCloser with additional functionality that can occur during read or close with the
	// expectation that the data is read from the up stream wrapped reader, and close operations are forwarded
	ConnectIn(r io.ReadCloser) PipeReader
}

// PipeTerm is a PipeWriter that performs no operation, useful for null-terminating a pipe line.
type PipeTerm struct{}

// Write for PipeTerm does nothing.
func (pt *PipeTerm) Write(b []byte) (n int, err error) {
	return len(b), nil
}

// ConnectOut for PipeTerm ignores the provided output stream and returns an empty
// PipeWriter.
func (pt *PipeTerm) ConnectOut(io.WriteCloser) PipeWriter {
	return pt
}

// Close for PipeTerm does nothing.
func (pt *PipeTerm) Close() error {
	return nil
}

// PipeCreateFile provides an output termination pipe segment that creates (or overwrites) a file for writing, and
// closes it when the pipeline is closed.
type PipeCreateFile struct {
	F *os.File
}

// NewPipeFileCreator creates a PipeCreateFile pipe segment. Internally, a new file with  the specified path is created
// or overwritten, ready for writing.
func NewPipeFileCreator(path string) (PipeWriter, error) {
	p := &PipeCreateFile{}
	var err error
	p.F, err = os.Create(path)
	if err != nil {
		return p, fmt.Errorf("error creating pipe file: %w", err)
	}
	return p, nil
}

// ConnectOut for PipeOut assigns the writecloser to the internal
// write closer.
func (p *PipeCreateFile) ConnectOut(w io.WriteCloser) PipeWriter {
	// p.W = w
	return p
}

// Write for PipeOut forwards writes to the internal write closer.
func (p *PipeCreateFile) Write(b []byte) (n int, err error) {
	if n, err = p.F.Write(b); err != nil {
		return n, fmt.Errorf("write in pipe create file: %w", err)
	}
	return n, nil
}

// Close for PipeOut forwards a close operation.
func (p *PipeCreateFile) Close() error {
	return p.F.Close()
}

// PipeOut is a PipeWriter wrapping around a standard write closer, forwarding all writes to the associated write
// closer.  This essentially provides a basic no-op forwarding implementation.
type PipeOut struct {
	W io.WriteCloser
}

// ConnectOut for PipeOut assigns the writecloser to the internal write closer.
func (p *PipeOut) ConnectOut(w io.WriteCloser) PipeWriter {
	p.W = w
	return p
}

// Write for PipeOut forwards writes to the internal write closer.
func (p *PipeOut) Write(b []byte) (n int, err error) {
	if n, err = p.W.Write(b); err != nil {
		return n, fmt.Errorf("write in pipe out: %w", err)
	}
	return n, nil
}

// Close for PipeOut forwards a close operation.
func (p *PipeOut) Close() error {
	return p.W.Close()
}

// PipeIn is a PipeReader wrapping around a standard read closer, forwarding all reads to the associated read closer.
// This essentially provides a basic no-op forwarding implementation.
type PipeIn struct {
	R io.ReadCloser
}

// ConnectIn for PipeIn assigns the readcloser to the internal read closer.
func (p *PipeIn) ConnectIn(r io.ReadCloser) PipeReader {
	p.R = r
	return p
}

// Read for PipeIn forwards reads to the internal read closer.
func (p *PipeIn) Read(d []byte) (n int, err error) {
	if n, err = p.R.Read(d); errors.Is(err, io.EOF) {
		return n, io.EOF
	} else if err != nil {
		return n, fmt.Errorf("read in pipe in: %w", err)
	}
	return n, nil
}

// Close for PipeIn forwards a close operation.
func (p *PipeIn) Close() error {
	return p.R.Close()
}

// PipeZstdEnc implements a PipeWriter that performs zstd encoding on
// the pipe stream.
type PipeZstdEnc struct {
	zstdw *zstd.Encoder
	w     io.WriteCloser
}

// ConnectOut for PipeZstdEnc wraps the incoming writecloser with a zstd encoder writer.
func (z *PipeZstdEnc) ConnectOut(w io.WriteCloser) PipeWriter {
	z.w = w
	zstdw, err := zstd.NewWriter(w)
	if err != nil {
		panic("zstd.NewWriter should never error")
	}
	z.zstdw = zstdw
	return z
}

// ConnectOutWithLevel for PipeZstdEnc wraps the incoming writecloser with a customized zstd encoder writer.
func (z *PipeZstdEnc) ConnectOutWithLevel(w io.WriteCloser, lv zstd.EncoderLevel) PipeWriter {
	z.w = w
	encOptions := zstd.WithEncoderLevel(lv)
	zstdw, err := zstd.NewWriter(w, encOptions)
	if err != nil {
		panic("zstd.NewWriter with level should never error")
	}
	z.zstdw = zstdw
	return z
}

// Write for PipeZstdEnc writes data to the zstd encoder (which in turn writes encoded data to the original output
// writer).
func (z *PipeZstdEnc) Write(p []byte) (n int, err error) {
	if n, err = z.zstdw.Write(p); err != nil {
		return n, fmt.Errorf("write in pipe zstd encoder: %w", err)
	}
	return n, nil
}

// Close for PipeZstdEnc closes the encoder.
func (z *PipeZstdEnc) Close() error {
	return errors.Join(z.zstdw.Close(), z.w.Close())
}

// NewPipeZstdEnc creates a PipeWriter that performs zstd compression on the passing stream.
func NewPipeZstdEnc() *PipeZstdEnc {
	z := &PipeZstdEnc{}
	return z
}

// AssignCompressionLevel Helper function to assign determine custom encoding level.
func AssignCompressionLevel(userSpecifiedLevel string) zstd.EncoderLevel {
	switch {
	case userSpecifiedLevel == "", strings.EqualFold(userSpecifiedLevel, "normal"):
		return zstd.SpeedDefault
	case strings.EqualFold(userSpecifiedLevel, "medium"):
		return zstd.SpeedBetterCompression
	case strings.EqualFold(userSpecifiedLevel, "max"), strings.EqualFold(userSpecifiedLevel, "maximum"):
		return zstd.SpeedBestCompression
	case strings.EqualFold(userSpecifiedLevel, "low"):
		return zstd.SpeedFastest
	default:
		return zstd.SpeedDefault
	}
}

// EncoderWithCustomLevel returns an Encoder at the specified compression level.
func EncoderWithCustomLevel(lv zstd.EncoderLevel) *zstd.Encoder {
	localEnc := zstd.Encoder{}
	// return an EOption with fastest level
	encoderLevel := zstd.EncoderLevel(lv)
	encoderOption := zstd.WithEncoderLevel(encoderLevel)

	z, err := zstd.NewWriter(&localEnc, encoderOption)
	if err != nil {
		panic("zstd.NewWriter should never error")
	}

	return z
}

// NewPipeZstdEncLevel creates a PipeWriter that performs zstd compression on the passing stream at the specified level.
func NewPipeZstdEncLevel(lv zstd.EncoderLevel) *PipeZstdEnc {
	z := &PipeZstdEnc{
		zstdw: EncoderWithCustomLevel(lv),
	}
	return z
}

// PipeZstdDec implements a PipeReader that performs zstd encoding on the pipe stream.
type PipeZstdDec struct {
	zstdr *zstd.Decoder
	r     io.Closer
}

// ConnectIn for PipeZstdDec wraps the incoming writecloser with a zstd encoder writer.
func (z *PipeZstdDec) ConnectIn(r io.ReadCloser) PipeReader {
	zstdr, err := zstd.NewReader(r)
	if err != nil {
		panic("zstd.NewReader should never error")
	}
	z.zstdr = zstdr
	z.r = r
	return z
}

// Read for PipeZstdDec reads data from the zstd decoder (which in turn reads encoded data from the original input
// reader).
func (z *PipeZstdDec) Read(p []byte) (n int, err error) {
	if n, err = z.zstdr.Read(p); errors.Is(err, io.EOF) {
		return n, io.EOF
	} else if err != nil {
		return n, fmt.Errorf("read in pipe zstd decoder: %w", err)
	}
	return n, nil
}

// Close for PipeZstdDec closes the encoder.
func (z *PipeZstdDec) Close() error {
	z.zstdr.Close()
	return z.r.Close()
}

// NewPipeZstdDec creates a PipeWriter that performs zstd decompression on the passing stream.
func NewPipeZstdDec() *PipeZstdDec {
	z := &PipeZstdDec{}
	return z
}

// PipeGzEnc implements a PipeWriter that performs zstd encoding on the pipe stream.
type PipeGzEnc struct {
	W io.WriteCloser
}

// Write for PipeGzEnc forwards writes to the internal write closer.
func (z *PipeGzEnc) Write(b []byte) (n int, err error) {
	if n, err = z.W.Write(b); err != nil {
		return n, fmt.Errorf("write in pipe gz encoder: %w", err)
	}
	return n, nil
}

// Close for PipeGzEnc forwards a close operation.
func (z *PipeGzEnc) Close() error {
	return z.W.Close()
}

// ConnectOut for PipeGzEnc wraps the incoming writecloser with a zstd encoder writer.
func (z *PipeGzEnc) ConnectOut(w io.WriteCloser) PipeWriter {
	ww, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
	if err != nil {
		panic("gzip.NewWriterLevel should never error")
	}
	z.W = ww
	return z
}

// NewPipeGzEnc creates a PipeWriter that performs zstd compression on the passing stream.
func NewPipeGzEnc() *PipeGzEnc {
	z := &PipeGzEnc{}
	return z
}

// PipeGzDec implements a PipeReader that performs gz encoding on the pipe stream.
type PipeGzDec struct {
	R io.ReadCloser
}

// ConnectIn for PipeGzDec wraps the incoming writecloser with a zstd encoder writer.
func (z *PipeGzDec) ConnectIn(r io.ReadCloser) PipeReader {
	rr, err := gzip.NewReader(r)
	if err != nil {
		panic("gzip.NewReader should never error")
	}
	z.R = rr
	return z
}

// Read for PipeGzDec forwards reads to the internal read closer.
func (z *PipeGzDec) Read(d []byte) (n int, err error) {
	if n, err = z.R.Read(d); errors.Is(err, io.EOF) {
		return n, io.EOF
	} else if err != nil {
		return n, fmt.Errorf("read in pipe gz decoder: %w", err)
	}
	return n, nil
}

// Close for PipeGzDec forwards a close operation.
func (z *PipeGzDec) Close() error {
	return z.R.Close()
}

// NewPipeGzDec creates a PipeWriter that performs gz decompression on the passing stream.
func NewPipeGzDec() *PipeGzDec {
	z := &PipeGzDec{}
	return z
}

// ProgressFunc is callback function used to update the number of bytes complete.
type ProgressFunc func(complete int64)

// PipeCounter is a PipeWriter that tracks the number of bytes passing through the stream.  This can be handy for
// monitoring sizes when other pipe stream segments modify the size of the data (eg due to compression).
type PipeCounter struct {
	W     io.WriteCloser
	R     io.ReadCloser
	PT    ProgressFunc
	Count int64
}

// ConnectOut for PipeCounter assigns the writecloser to the internal write closer.
func (pc *PipeCounter) ConnectOut(w io.WriteCloser) PipeWriter {
	pc.W = w
	return pc
}

// ConnectIn for PipeCounter assigns the readcloser to the internal read closer.
func (pc *PipeCounter) ConnectIn(r io.ReadCloser) PipeReader {
	pc.R = r
	return pc
}

// AddProgressTracker for PipeCounter adds progress tracking functionality to the pipeline.
func (pc *PipeCounter) AddProgressTracker(pt ProgressFunc) PipeReader {
	pc.PT = pt
	return pc
}

// Write for PipeCounter increments the counter, and forwards the output to the internal writer.
func (pc *PipeCounter) Write(p []byte) (n int, err error) {
	pc.Count += int64(len(p))
	if pc.PT != nil {
		pc.PT(int64(len(p)))
	}
	if n, err = pc.W.Write(p); err != nil {
		return n, fmt.Errorf("write in pipe counter: %w", err)
	}
	return n, nil
}

// Read for PipeCounter increments the counter, and forwards the output to the internal reader.
func (pc *PipeCounter) Read(p []byte) (n int, err error) {
	pc.Count += int64(len(p))
	if pc.PT != nil {
		pc.PT(int64(len(p)))
	}
	if n, err = pc.R.Read(p); errors.Is(err, io.EOF) {
		return n, io.EOF
	} else if err != nil {
		return n, fmt.Errorf("read in pipe counter: %w", err)
	}
	return n, nil
}

// Close for PipeCounter forwards the close action to either the read or write streams if they are defined.
func (pc *PipeCounter) Close() error {
	if pc.W != nil {
		return pc.W.Close()
	}
	if pc.R != nil {
		return pc.R.Close()
	}
	return nil
}

// NewPipeCounter returns a fresh PipeCounter ready for connection.
func NewPipeCounter() *PipeCounter {
	return &PipeCounter{}
}

// PipeDigest performs a Sha256 digest on the passing stream.  This is done in quasi-parallel, writing unbuffered
// stream to the digester then to the connected writer
// Warning: retrieving the digest hash while a write is in progress can cause a panic.  It is best to ensure the
// pipeline is closed (by calling Close() on the 'entrance' segment) before checking the digest.
type PipeDigest struct {
	W        io.WriteCloser
	R        io.ReadCloser
	Digester hash.Hash
}

// NewPipeDigest creates a new pipewriter segment that performs a digest of the passing stream, ready for connection.
func NewPipeDigest() *PipeDigest {
	// TODO make this digest algorithm configurable
	return &PipeDigest{Digester: sha256.New()}
}

// ConnectOut for PipeDigest assigns the writecloser to the internal write closer.
func (pd *PipeDigest) ConnectOut(w io.WriteCloser) PipeWriter {
	pd.W = w
	return pd
}

// ConnectIn for PipeDigest assigns the readcloser to the internal read closer.
func (pd *PipeDigest) ConnectIn(r io.ReadCloser) PipeReader {
	pd.R = r
	return pd
}

// Write for PipeDigest calculates the digest by writing to the digester each chunk of bytes is written to the digester
// and then to the connected writer.
func (pd *PipeDigest) Write(p []byte) (n int, err error) {
	n, err = pd.Digester.Write(p)
	if err != nil {
		return
	}
	if n, err = pd.W.Write(p); err != nil {
		return n, fmt.Errorf("write in pipe digest: %w", err)
	}
	return n, nil
}

// Read for PipeDigest calculates the digest by writing to the digester each chunk of bytes is read from the reader,
// and then written  to the digester for the calculation.
func (pd *PipeDigest) Read(p []byte) (n int, err error) {
	n, err = pd.R.Read(p)
	if errors.Is(err, io.EOF) {
		return n, io.EOF
	}
	if err != nil {
		return n, fmt.Errorf("read in pipe digest: %w", err)
	}
	if n, err = pd.Digester.Write(p); err != nil {
		return n, fmt.Errorf("write in pipe digest: %w", err)
	}
	return n, nil
}

// Close for PipeDigest forwards the close action to either the read or write streams if they are defined.
func (pd *PipeDigest) Close() error {
	if pd.W != nil {
		return pd.W.Close()
	}
	if pd.R != nil {
		return pd.R.Close()
	}
	return nil
}

// GetDigest returns the current digest stored in the digest stream writer.
func (pd *PipeDigest) GetDigest() digest.Digest {
	return digest.NewDigestFromBytes(digest.SHA256, pd.Digester.Sum(nil))
}

// BufWriteCloser is a wrapper around bytes.buffer that provides a no-op close for io.WriteCloser compatibility.
type BufWriteCloser struct {
	*bytes.Buffer
}

// Close for BufWriteCloser doesn't need to do anything, just for interface compat.
func (bwc *BufWriteCloser) Close() error {
	return nil
}

// FlushWriter combines closer, writer, and Flush, used for tracking compressors that must flush data before checking
// the output size.
type FlushWriter interface {
	io.Writer
	io.Closer
	Flush() error
}

// PipeCompressThreshold performs compression on input and forwards compressed data to the output, as long as the
// compression ratio meets a specified threshold.  If the compression threshold is NOT met, the data is instead piped
// directly to the output, uncompressed.
// To perform this, the first N bytes of the input are buffered, then compressed into a second buffer.  When the data
// size reaches N, the size of the two buffers are compared to determine the ratio.  If the ratio meets desired
// threshold, the compressed buffer is forwarded to the output pipe, and incoming data is no longer buffered.
// Conversely, if the ratio is not met, the uncompressed buffer is forwarded to the output, and the compression process
// is bypassed for the remainder of the data.
type PipeCompressThreshold struct {
	W io.WriteCloser // final output writer

	bufW   io.Writer      // a temporary buffered writer to catch data until confirmation is made
	uncBuf BufWriteCloser // uncompressed temporary buffer
	cmpBuf BufWriteCloser // compressed temporary buffer

	certain   bool // whether we have come to a compression decision -- false means the decision is still pending
	bufCount  int  // count of bytes written to the buffer, when this exceeds checkSize a size comparison is done
	checkSize int  // size of bytes to use for compression comparison.  Two buffers of up to this size are created in mem

	compressor    FlushWriter       // the compressor used for compression
	compressLevel zstd.EncoderLevel // compression level setting, ignored for gz
	compressGz    bool              // true to use a gzip compressor, false for zstd

	FinalRatio  float32 // if nonzero, this is the resulting ratio  (uncompressed/compressed) for reporting purposes
	DidCompress bool    // true if the compression ratio check passed and the data was compressed.
}

// NewPipeCompressThreshold creates a pipe writer that will buffer the initial checkSize bytes, keeping two versions,
// a compressed and uncompressed one.  When the checkSize is met, the compressibility of the data is examined.  If the
// data is not compressible, then the uncompressed data is output.  Further writes after the  check will no
// longer buffer data, and will instead pipe data directly to the output (or compressor).
// lvl indicates a zstd compression level to use.  Currently, levels are only supported for zstd, and lvl is ignored for
// Gzip. useGz should be set to true to use a gzip encoder versus the zstd one.
func NewPipeCompressThreshold(lvl zstd.EncoderLevel, checkSize int, useGz bool) *PipeCompressThreshold {
	return &PipeCompressThreshold{compressLevel: lvl, compressGz: useGz, checkSize: checkSize}
}

// Close for PipeCompressThreshold will simply forward a close operation down the pipe if the compression check
// threshold has already been met.  If that threshold has not been met, however, we have to flush the compressed or
// uncompressed buffer to the output based on compression ratio.
func (pct *PipeCompressThreshold) Close() error {
	// If we are certain, we've already passed the threshold and we don't need to do any further special handling for
	// buffered data.
	if pct.certain {
		return pct.W.Close()
	}
	// We're uncertain, so check our compression ratio and either output the compressed or uncompressed data
	if pct.IsCompressible() {
		// copy compressed data to the output
		_, _ = io.Copy(pct.W, pct.cmpBuf)
		// note: we don't create a new compressor here because we're terminating
	} else {
		// copy uncompressed data to the output
		_, _ = io.Copy(pct.W, pct.uncBuf)
	}
	return pct.W.Close()
}

// ConnectOut establishes the temporary buffers, and sets the final output for when a compression decision has been
// reached. Until the threshold is reached, data is multiwritten to two temporary buffered outputs.  Due to this
// buffering, output from PipeCompressThreshold is delayed (the delay goes away once the threshold is met).  Ultimately,
// the data will be either compressed or uncompressed depending on the threshold calculation.
func (pct *PipeCompressThreshold) ConnectOut(w io.WriteCloser) PipeWriter {
	pct.uncBuf = BufWriteCloser{Buffer: new(bytes.Buffer)}
	pct.cmpBuf = BufWriteCloser{Buffer: new(bytes.Buffer)}

	pct.compressor = pct.createCompressor(pct.cmpBuf)
	pct.bufW = io.MultiWriter(pct.uncBuf, pct.compressor)
	pct.W = w

	return pct
}

// createCompressor generates a new compression encoder, either zstd or gzip depending on settings, and according to
// the configured compression level.
func (pct *PipeCompressThreshold) createCompressor(writer io.Writer) FlushWriter {
	if pct.compressGz {
		// currently we don't support alternate gzip compression levels
		ww, err := gzip.NewWriterLevel(writer, gzip.DefaultCompression)
		if err != nil {
			panic("gzip.NewWriterLevel should never error")
		}
		return ww
	}
	encOptions := zstd.WithEncoderLevel(pct.compressLevel)
	zstdw, err := zstd.NewWriter(writer, encOptions)
	if err != nil {
		panic("zstd.NewWriter with level should never error")
	}
	return zstdw
}

// Write for PipeCompressThreshold writes incoming data to two output streams.  The first is an uncompressed buffer,
// the second is a compressor to a compressed buffer.  If checkThreshold is reached, the buffer sizes are compared,
// and the rest of the data is forwarded only to the final output stream.
func (pct *PipeCompressThreshold) Write(p []byte) (n int, err error) {
	// if we've made a compression determination, don't write to the buffers anymore, send directly to either the
	// output or the compressor
	if pct.certain {
		n, err = pct.W.Write(p)
		return
	}

	// write the latest packet to the buffers.  This is a multiwriter that sends to uncompressed buffer and compressor
	n, err = pct.bufW.Write(p)

	// accumulate the total bytes written
	pct.bufCount += n

	// check against our check threshold. If we've written enough, lets check compressibility
	if pct.bufCount >= pct.checkSize {
		// check compression ratio based on current size of compressed and uncompressed buffer
		if pct.IsCompressible() {
			// data is at least nominally compressible
			// create a new compressor with the true output, and set it up as the destination for further writes
			pct.W = pct.createCompressor(pct.W)
			// Can't reliably swap the output as the compressor was flushed, which leads to a non-identical compression
			// result if we wish to continue.  Instead, we recompress from the uncompressed buffer before continuing
			_, _ = io.Copy(pct.W, pct.uncBuf)
		} else {
			// un-compressible, copy unmodified data from buffer to output.  Further writes will go directly to the
			// output
			_, _ = io.Copy(pct.W, pct.uncBuf)
		}
		// mark ourselves as certain, to shortcut the buffering process in the future
		pct.certain = true
	}
	return
}

// IsCompressible returns true if the ratio compressed/uncompressed is less than a threshold.  This is safe to call
// after the write process is complete to determine if the initial data was compressible.
func (pct *PipeCompressThreshold) IsCompressible() bool {
	if pct.FinalRatio != 0 {
		return pct.DidCompress
	}
	if pct.compressor != nil {
		if err := pct.compressor.Close(); err != nil {
			panic("compressor close failure to buffer")
		}
	}
	if pct.cmpBuf.Len() == 0 {
		return false
	}
	// see https://en.wikipedia.org/wiki/Data_compression_ratio
	pct.FinalRatio = float32(pct.uncBuf.Len()) / float32(pct.cmpBuf.Len())
	pct.DidCompress = pct.FinalRatio >= compressionRatioThreshold
	return pct.DidCompress
}
