package bottle

import (
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
)

// PushOptions provides options for pushing bottles from the localhost to a remote registry.
type PushOptions struct {
	tbottle.TransferOptions
}
