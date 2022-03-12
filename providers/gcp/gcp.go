package gcp

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"github.com/stevequadros/uploader/config"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io"
	"os"
)

type Uploader struct {
	credentials *google.Credentials
	client      *storage.Client
}

func New(ctx context.Context, config *config.GCPConfig) (*Uploader, error) {
	if config.Credentials == nil {
		return nil, errors.New("gcp credentials are empty")
	}

	f, err := os.Open(config.Credentials.Filename)
	if err != nil {
		return nil, err
	}
	credentialsJson, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	credentials, err := google.CredentialsFromJSON(ctx, credentialsJson, config.Credentials.Scopes...)
	if err != nil {
		return nil, err
	}
	opts := option.WithCredentials(credentials)
	client, err := storage.NewClient(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Uploader{credentials: credentials, client: client}, nil
}

func (u *Uploader) Upload(ctx context.Context, bucketName, key string, file *os.File) error {
	bucket := u.client.Bucket(bucketName)
	if err := bucket.Create(ctx, u.credentials.ProjectID, &storage.BucketAttrs{}); err != nil {
		return err
	}
	obj := bucket.Object(key)
	writer := obj.NewWriter(ctx)
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	if err != nil {
		return err
	}
	return writer.Close()
}
