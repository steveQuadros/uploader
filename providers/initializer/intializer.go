package initializer

import (
	"context"
	"errors"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"github.com/stevequadros/uploader/providers/aws"
	"github.com/stevequadros/uploader/providers/azure"
	"github.com/stevequadros/uploader/providers/gcp"
)

func Init(ctx context.Context, config config.Config) ([]providers.Uploader, error) {
	var provs []providers.Provider
	if config.AWS != nil {
		provs = append(provs, providers.AWS)
	}
	if config.Azure != nil {
		provs = append(provs, providers.Azure)
	}
	if config.GCP != nil {
		provs = append(provs, providers.GCP)
	}

	var uploaders []providers.Uploader
	for _, p := range provs {
		switch p {
		case providers.AWS:
			u, err := initAWS(config.AWS)
			if err != nil {
				return uploaders, err
			}
			uploaders = append(uploaders, u)
		case providers.GCP:
			u, err := initGCP(ctx, config.GCP)
			if err != nil {
				return uploaders, err
			}
			uploaders = append(uploaders, u)
		case providers.Azure:
			u, err := initAzure(config.Azure)
			if err != nil {
				return uploaders, err
			}
			uploaders = append(uploaders, u)
		default:
			return nil, errors.New("unknown provider")
		}
	}
	return uploaders, nil
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
