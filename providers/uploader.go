package providers

import (
	"context"
	"fmt"
	"io"
)

type Provider string

const (
	AWS   Provider = "aws"
	GCP   Provider = "gcp"
	Azure Provider = "azure"
)

var Providers = map[Provider]struct{}{
	AWS:   {},
	GCP:   {},
	Azure: {},
}

type Uploader interface {
	Upload(ctx context.Context, bucket, key string, reader io.ReadSeekCloser) error
}

type UploadError struct {
	provider Provider
	err      error
}

var _ error = (*UploadError)(nil)

func (e UploadError) Error() string {
	return fmt.Sprintf("%q upload failed, error: %q", string(e.provider), e.err.Error())
}

func NewUploadError(provider Provider, err error) UploadError {
	return UploadError{provider: provider, err: err}
}
