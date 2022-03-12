package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/oauth2/google"
	"os"
)

type Uploader struct {
	credentials *google.Credentials
	client      *s3manager.Uploader
}

func New() (*Uploader, error) {
	// The session the S3 Uploader will use
	provider := &credentials.SharedCredentialsProvider{
		Filename: "",
		//Profile:  "default",
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
	return &Uploader{
		client: client,
	}, nil
}

func (u *Uploader) Upload(ctx context.Context, bucket, key string, file *os.File) error {
	_, err := u.client.S3.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}
	// Upload the file to S3.
	_, err = u.client.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	return nil
}
