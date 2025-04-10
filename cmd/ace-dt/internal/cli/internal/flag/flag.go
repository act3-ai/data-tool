// Package flag defines a command flag for specifying a telemetry URL.
package flag

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/pflag"

	"github.com/act3-ai/data-tool/internal/actions"
)

// TelemetryURLFlags adds a flag for changing the default telemetry url.
func TelemetryURLFlags(flags *pflag.FlagSet, action *actions.TelemetryOptions) {
	flags.StringVar(&action.URL, "telemetry", "",
		`Overrides the telemetry server configuration with the single telemetry server URL provided.  
Modify the configuration file if multiple telemetry servers should be used or if auth is required.`)
}

// BytesValue is a pflags.Value compatible that allows for passing byte values easily.
// It supports SI suffixes for bytes.
type BytesValue uint64

// Set implements the flag.Value and pflag.Value interfaces.
func (b *BytesValue) Set(s string) error {
	v, err := humanize.ParseBytes(s)
	if err != nil {
		return fmt.Errorf("parsing command line: %w", err)
	}
	*b = BytesValue(v)
	return nil
}

// Type implements the pflag.Value interface.
func (b *BytesValue) Type() string {
	return "bytes"
}

// String implements the flag.Value and pflag.Value interfaces.
func (b *BytesValue) String() string {
	return fmt.Sprintf("%s (%d B)", humanize.Bytes(uint64(*b)), *b)
}

// MemoryBufferOptions are the options for mbuffer (aka BlockBuf).
type MemoryBufferOptions struct {
	// Size is the total size of the memory buffer
	Size BytesValue

	// BlockSize is the block size
	BlockSize BytesValue

	// HighWaterMarkPercentage is the high water mark as a percentage
	HighWaterMarkPercentage int
}

// Options will convert the options values and return
// the number of blocks, block size (B), number of blocks to reach the HWM.
func (opts *MemoryBufferOptions) Options() (int, int, int) {
	// convert to number of blocks
	n := iceil(int(opts.Size), int(opts.BlockSize))
	hwm := iceil(int(opts.Size)*opts.HighWaterMarkPercentage/100, int(opts.BlockSize))

	return n, int(opts.BlockSize), hwm
}

// iceil return the ceiling of x divided by y using only integer arithmetic.
func iceil(x, y int) int {
	return 1 + (x-1)/y
}

// AddMemoryBufferFlags adds flags for the memoryBufferOptions to the flag set.
func AddMemoryBufferFlags(flags *pflag.FlagSet, options *MemoryBufferOptions) {
	flags.VarP(&options.Size, "buffer-size", "m", "Size of the memory buffer. Si suffixes are supported.")
	flags.VarP(&options.BlockSize, "block-size", "b", "Block size used for writes.  Si suffixes are supported.")
	flags.IntVar(&options.HighWaterMarkPercentage, "hwm", 90, "Percentage of buffer to fill before writing")
}
