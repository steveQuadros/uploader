package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stevequadros/uploader/config"
	"os"
)

type Uploader struct {
	client *azblob.ServiceClient
}

func New(config *config.AzureConfig) (*Uploader, error) {
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
	return &Uploader{client: &serviceClient}, nil
}

func (u *Uploader) Upload(ctx context.Context, bucket, key string, file *os.File) error {
	containerClient := u.client.NewContainerClient(bucket)
	_, err := containerClient.Create(ctx, &azblob.CreateContainerOptions{})
	if err != nil {
		return err
	}
	blobClient := containerClient.NewBlockBlobClient(key)
	_, err = blobClient.Upload(ctx, file, &azblob.UploadBlockBlobOptions{})
	if err != nil {
		return err
	}
	return nil
}
