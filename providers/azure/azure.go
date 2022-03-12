package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"io"
)

type AzureUploader struct {
	client *azblob.ServiceClient
}

var _ providers.Uploader = (*AzureUploader)(nil)

func New(config config.AzureConfig) (*AzureUploader, error) {
	if config.Credentials == nil {
		return nil, errors.New("azure credentials are empty")
	}

	connStr := fmt.Sprintf(
		"DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net",
		config.Credentials.AccountName,
		config.Credentials.AccountKey,
	)
	serviceClient, err := azblob.NewServiceClientFromConnectionString(connStr, nil)
	if err != nil {
		return nil, err
	}
	return &AzureUploader{client: &serviceClient}, nil
}

func (u *AzureUploader) Upload(ctx context.Context, bucket, key string, reader io.ReadSeekCloser) error {
	containerClient := u.client.NewContainerClient(bucket)
	_, err := containerClient.Create(ctx, &azblob.CreateContainerOptions{})
	if err != nil {
		return err
	}
	blobClient := containerClient.NewBlockBlobClient(key)
	_, err = blobClient.Upload(ctx, reader, &azblob.UploadBlockBlobOptions{})
	if err != nil {
		return providers.NewUploadError("Azure", err)
	}
	return nil
}
