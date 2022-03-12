package main

import (
	"context"
	"fmt"
	"github.com/stevequadros/uploader/providers/aws"
	"log"
	"os"
)

// providers
// autocreate bucket default to false, prompt if bucket not exists
// Gcp complains if you try to create an existing bucket

func main() {
	file, err := os.Open("test.txt")
	if err != nil {
		log.Fatal("could not open file", err)
	}

	ctx := context.Background()
	//client, err := gcp.New(ctx, "/Users/stevenquadros/Downloads/files-com-uploader-cfcef445bead.json")
	//if err != nil {
	//	log.Fatal("error opening client", err)
	//}
	//if err = client.Upload(ctx, fmt.Sprintf("%s-%s", "filescom-takehome-", "test"), file.Name(), file); err != nil {
	//	log.Fatal("error uploading file", err)
	//}

	awsClient, err := aws.New()
	if err != nil {
		log.Fatal("error creating aws", err)
	}
	if err = awsClient.Upload(ctx, fmt.Sprintf("%s-%s", "filescom-takehome-", "test"), file.Name(), file); err != nil {
		log.Fatal("error uploading to s3, ", err)
	}
}
