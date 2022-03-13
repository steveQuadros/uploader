package coordinator

import (
	"context"
	"errors"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"github.com/stevequadros/uploader/providers/aws"
	"github.com/stevequadros/uploader/providers/azure"
	"github.com/stevequadros/uploader/providers/gcp"
	"io"
	"sync"
)

type Coordinator struct {
	config    config.Config
	uploaders []providers.Uploader
	providers []providers.Provider
}

func NewCoordinator(ctx context.Context, cfg config.Config) (Coordinator, error) {
	var provs []providers.Provider
	if cfg.AWS != nil {
		provs = append(provs, providers.AWS)
	}
	if cfg.Azure != nil {
		provs = append(provs, providers.Azure)
	}
	if cfg.GCP != nil {
		provs = append(provs, providers.GCP)
	}

	c := Coordinator{config: cfg, providers: provs}
	for _, p := range c.providers {
		uploader, err := c.initProvider(ctx, p)
		if err != nil {
			return c, err
		}
		c.uploaders = append(c.uploaders, uploader)
	}

	return c, nil
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
	uploadErrors := make(chan DoError, len(c.providers))
	success := make(chan providers.Provider, len(c.providers))
	wg := sync.WaitGroup{}

	var count int
	for i, u := range c.uploaders {
		wg.Add(1)
		go func(client providers.Uploader, n int) {
			p := c.providers[n]
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
	for count < len(c.providers) {
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

	if len(failed) == len(c.providers) {
		return doResult, errors.New("all uploads Failed")
	} else if len(failed) > 0 {
		return doResult, errors.New("some uploads Failed")
	} else {
		return doResult, nil
	}
}

func (c *Coordinator) Providers() []providers.Provider {
	return c.providers
}

func (c *Coordinator) initProvider(ctx context.Context, p providers.Provider) (providers.Uploader, error) {
	switch p {
	case providers.AWS:
		return initAWS(c.config.AWS)
	case providers.GCP:
		return initGCP(ctx, c.config.GCP)
	case providers.Azure:
		return initAzure(c.config.Azure)
	default:
		return nil, errors.New("unknown provider")
	}
}

func initAWS(cfg *config.AWS) (*aws.AWSUploader, error) {
	client, err := aws.New(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initGCP(ctx context.Context, cfg *config.GCP) (*gcp.GCPUploader, error) {
	client, err := gcp.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initAzure(cfg *config.Azure) (*azure.AzureUploader, error) {
	client, err := azure.New(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
