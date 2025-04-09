package bottle

import (
	tbottle "github.com/act3-ai/data-tool/pkg/transfer/bottle"
)

// PushOptions provides options for pushing bottles from the localhost to a remote registry.
type PushOptions struct {
	tbottle.TransferOptions
}
