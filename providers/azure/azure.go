package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"io"
	"io/ioutil"
	"os"
)

type AzureUploader struct {
	client *azblob.ServiceClient
}

var _ providers.Uploader = (*AzureUploader)(nil)

func New(config *config.Azure) (*AzureUploader, error) {
	if config == nil || config.Credentials == nil {
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

func (u *AzureUploader) GetName() providers.Provider {
	return providers.Azure
}

func (u *AzureUploader) Upload(ctx context.Context, bucket, key string, reader io.ReadSeekCloser) error {
	containerClient := u.client.NewContainerClient(bucket)
	// if container already exists, proceed without creation
	res, err := containerClient.GetProperties(ctx, nil)
	if res.ETag == nil {
		_, err = containerClient.Create(ctx, &azblob.CreateContainerOptions{})
		if err != nil {
			return err
		}
	}

	// azure closes the file, which will cause future calls to fail
	// this is a workaround for now, but we'll copy it over to avoid this - in real application
	// I'd prefer a better method since this could cause issues with large files
	// NOTE - azureblob.Upload closes file, so no need to call
	tmp, err := ioutil.TempFile("", "filescom-tmp-")
	if err != nil {
		return err
	}
	defer func() {
		os.Remove(tmp.Name())
	}()

	_, err = io.Copy(tmp, reader)
	if err != nil {
		return err
	}

	// reset file to start after copy
	_, err = tmp.Seek(0, 0)
	if err != nil {
		return err
	}

	blobClient := containerClient.NewBlockBlobClient(key)
	_, err = blobClient.Upload(ctx, tmp, &azblob.UploadBlockBlobOptions{})
	if err != nil {
		return providers.NewUploadError("Azure", err)
	}
	return nil
}
