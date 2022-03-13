package aws

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"golang.org/x/oauth2/google"
	"io"
)

type AWSUploader struct {
	credentials *google.Credentials
	client      *s3manager.Uploader
}

var _ providers.Uploader = (*AWSUploader)(nil)

func New(config config.AWS) (*AWSUploader, error) {
	if config.Credentials == nil {
		return nil, errors.New("AWS credentials are empty")
	}
	// The session the S3 AWSUploader will use
	provider := &credentials.SharedCredentialsProvider{
		Filename: config.Credentials.Filename,
		Profile:  config.Credentials.Profile,
	}
	creds := credentials.NewCredentials(provider)
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		return nil, err
	}

	// Create an uploader with the session and default options
	client := s3manager.NewUploader(sess)
	return &AWSUploader{
		client: client,
	}, nil
}

func (u *AWSUploader) Upload(ctx context.Context, bucket, key string, reader io.ReadSeekCloser) error {
	_, err := u.client.S3.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}
	// Upload the file to S3.
	_, err = u.client.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return providers.NewUploadError("AWS", err)
	}
	return nil
}
