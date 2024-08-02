package mirror

import "context"

type BatchSerialize struct {
	*Action
}

func (action *BatchSerialize) Run(ctx context.Context, trackerFile, syncDir string) error {

	return nil
}
