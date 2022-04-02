package destination

import (
	"context"
	"io"
)

type Target interface {
	Upload(ctx context.Context, data io.Reader) error
}
