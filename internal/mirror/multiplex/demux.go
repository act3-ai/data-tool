package multiplex

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Demux reads the multiplexed stream from r and writes to the corresponding writer.
func Demux(r io.Reader, writers ...io.Writer) error {
	var buf []byte
	var headerBuf [headerSize]byte

	for {
		// read the header data
		if _, er := io.ReadFull(r, headerBuf[:]); er != nil {
			if errors.Is(er, io.EOF) {
				break
			}
			return fmt.Errorf("reading header: %w", er)
		}
		// decode the header
		id := binary.LittleEndian.Uint32(headerBuf[:4])
		size := binary.LittleEndian.Uint64(headerBuf[4:])

		// reuse the data buffer if possible
		if buf == nil || len(buf) < int(size) {
			buf = make([]byte, size)
		}
		data := buf[:size]

		// read the data
		if _, er := io.ReadFull(r, data); er != nil {
			return fmt.Errorf("reading body: %w", er)
		}

		// write the data to the appropriate writer
		if _, err := writers[id].Write(data); err != nil {
			return fmt.Errorf("writing to %d failed: %w", id, err)
		}
	}
	return nil
}
