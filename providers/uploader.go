package providers

import (
	"context"
	"io"
)

type Uploader interface {
	Upload(ctx context.Context, bucket, key string, reader io.ReadSeekCloser) error
}
