package source

import (
	"context"
)

type Provider interface {
	LocalURI(ctx context.Context, sourceURL string) (string, error)
}
