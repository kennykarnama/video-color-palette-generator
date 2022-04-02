package destination

import (
	"io"
	"context"
)

type noop struct{}

func NewNoop(foo string) *noop {
	return &noop{}
}

func (n *noop) Upload(ctx context.Context, data io.Reader) error {
	return nil
}
