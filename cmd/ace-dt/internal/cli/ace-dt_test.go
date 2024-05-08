package cli

import (
	"github.com/spf13/cobra"
)

func rootTestCmd() *cobra.Command {
	return NewToolCmd("0.0.0")
}
