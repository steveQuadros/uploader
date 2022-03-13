package coordinator

import (
	"context"
	"errors"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"io"
	"sync"
)

type Coordinator struct {
	config    config.Config
	uploaders []providers.Uploader
}

func NewCoordinator(uploaders []providers.Uploader) (Coordinator, error) {
	return Coordinator{
		uploaders: uploaders,
	}, nil
}

type DoError struct {
	Provider providers.Provider
	Error    error
}

type DoResult struct {
	Done   []providers.Provider
	Failed []DoError
}

func (c *Coordinator) Do(ctx context.Context, bucket, key string, reader io.ReadSeekCloser) (DoResult, error) {
	uploadErrors := make(chan DoError, len(c.uploaders))
	success := make(chan providers.Provider, len(c.uploaders))
	wg := sync.WaitGroup{}

	var count int
	for i, u := range c.uploaders {
		wg.Add(1)
		go func(client providers.Uploader, n int) {
			p := u.GetName()
			if uploadErr := client.Upload(ctx, bucket, key, reader); uploadErr != nil {
				uploadErrors <- DoError{p, uploadErr}
			} else {
				success <- p
			}
			wg.Done()
		}(u, i)
	}

	var done []providers.Provider
	var failed []DoError
	for count < len(c.uploaders) {
		select {
		case e := <-uploadErrors:
			failed = append(failed, e)
			count++
		case p := <-success:
			done = append(done, p)
			count++
		}
	}

	wg.Wait()
	close(uploadErrors)
	close(success)
	doResult := DoResult{
		Done:   done,
		Failed: failed,
	}

	if len(failed) == len(c.uploaders) {
		return doResult, errors.New("all uploads Failed")
	} else if len(failed) > 0 {
		return doResult, errors.New("some uploads Failed")
	} else {
		return doResult, nil
	}
}
