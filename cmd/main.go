package main

import (
	"context"
	"fmt"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers/aws"
	"github.com/stevequadros/uploader/providers/azure"
	"github.com/stevequadros/uploader/providers/gcp"
	"log"
	"os"
)

// providers
// autocreate bucket default to false, prompt if bucket not exists
// Gcp complains if you try to create an existing bucket
// - move bucket creation to interface function for easy testing of paths
// - move uploader behind interface for easy testing`

const (
	AzureAccountName string = "AZURE_ACCOUNT_NAME"
	AzureKey         string = "AZURE_KEY"
	AWSFilename      string = "AWS_FILENAME"
	AWSProfile       string = "AWS_PROFILE"
	GCPFilename      string = "GCP_FILENAME"
)

func main() {
	file, err := os.Open("test.txt")
	if err != nil {
		log.Fatal("could not open file", err)
	}

	ctx := context.Background()
	gcpFilename := os.Getenv(GCPFilename)
	if gcpFilename == "" {
		log.Fatal("gcp filename is not set")
	}
	client, err := gcp.New(ctx, config.NewGCP(gcpFilename))
	if err != nil {
		log.Fatal("error opening client", err)
	}
	if err = client.Upload(ctx, fmt.Sprintf("%s-%s", "filescom-takehome-", "test"), file.Name(), file); err != nil {
		log.Fatal("error uploading file", err)
	}

	awsFilename := os.Getenv(AWSFilename)
	if awsFilename == "" {
		log.Fatal("aws filename is empty")
	}
	awsProfile := os.Getenv(AWSProfile)
	if awsProfile == "" {
		log.Fatal("aws profile is empty")
	}
	awsClient, err := aws.New(config.NewAWS(awsFilename, awsProfile))
	if err != nil {
		log.Fatal("error creating aws", err)
	}
	if err = awsClient.Upload(ctx, fmt.Sprintf("%s-%s", "filescom-takehome-", "test"), file.Name(), file); err != nil {
		log.Fatal("error uploading to s3, ", err)
	}

	azureAccountName := os.Getenv(AzureAccountName)
	if azureAccountName == "" {
		log.Fatal("azure account name is empty")
	}
	azureKey := os.Getenv(AzureKey)
	if azureKey == "" {
		log.Fatal("azure key is empty")
	}
	azureClient, err := azure.New(config.NewAzure(azureAccountName, azureKey))
	if err != nil {
		log.Fatal("error creating azure", err)
	}
	if err = azureClient.Upload(ctx, "filescom-takehome", file.Name(), file); err != nil {
		log.Fatal("error uploading azure", err)
	}
}
