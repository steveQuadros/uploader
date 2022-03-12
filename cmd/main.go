package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/stevequadros/uploader/config"
	xproviders "github.com/stevequadros/uploader/providers"
	"github.com/stevequadros/uploader/providers/aws"
	"github.com/stevequadros/uploader/providers/azure"
	"github.com/stevequadros/uploader/providers/gcp"
	"log"
	"os"
	"strings"
	"sync"
)

// providers
// autocreate bucket default to false, prompt if bucket not exists
// Gcp complains if you try to create an existing bucket TOP PRIORITY
// - move bucket creation to interface function for easy testing of paths
// - check all buckets are valid before trying to upload to avoid partial uploads TOP
// - move uploader behind interface for easy testing`
// verify file was uploaded code for ease of use - maybe
// uploading files should attempt all and return list of errors rather than fast failing
// os.Getenv for additional option to config file
// configs need proper json annotations

type Provider string

const (
	AzureAccountName string   = "AZURE_ACCOUNT_NAME"
	AzureKey         string   = "AZURE_KEY"
	AWSFilename      string   = "AWS_FILENAME"
	AWSProfile       string   = "AWS_PROFILE"
	GCPFilename      string   = "GCP_FILENAME"
	AWS              Provider = "aws"
	GCP              Provider = "gcp"
	Azure            Provider = "azure"
)

var ValidProviders = map[Provider]struct{}{
	AWS:   {},
	GCP:   {},
	Azure: {},
}

type providerFlag []Provider

func (i *providerFlag) String() string {
	b := strings.Builder{}
	for _, p := range *i {
		b.WriteString(string(p))
	}
	return b.String()
}

func (i *providerFlag) Set(value string) error {
	*i = append(*i, Provider(value))
	return nil
}

func main() {
	var providers providerFlag
	var filename, configPath, bucket, key string
	flag.Var(&providers, "provider", "Providers targeted. Valid Options: aws, gcp, azure. Each one must be preceded with it's own flag, ex: -provider aws -provider azure -provider gcp")
	flag.StringVar(&filename, "file", "", "The file to upload")
	flag.StringVar(&configPath, "config", "", "Path to config.json")
	flag.StringVar(&bucket, "bucket", "", "target bucket for file")
	flag.StringVar(&key, "key", "", "key for file")
	flag.Parse()

	if err := validateProviders(providers); err != nil {
		log.Fatal("error validating providers: ", err)
	}

	if filename == "" {
		log.Fatal("no file provided")
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("could not open file", err)
	}

	if configPath == "" {
		log.Fatal("no config filepath passed")
	}
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal("error opening config", err)
	}
	cfg, err := config.NewFromJSON(configFile)
	if err != nil {
		log.Fatal("error parsing configfile", err)
	}

	if bucket == "" {
		log.Fatal("no bucket provided")
	}

	if key == "" {
		log.Fatal("no key provided")
	}

	ctx := context.Background()

	var clients []xproviders.Uploader
	var initErrors []error
	for _, p := range providers {
		client, err := initProvider(ctx, p, cfg)
		if err != nil {
			initErrors = append(initErrors, err)
		}
		clients = append(clients, client)
	}

	if len(initErrors) != 0 {
		for _, e := range initErrors {
			fmt.Println(e)
		}
		return
	}
	uploadErrors := make(chan error, len(providers))
	wg := sync.WaitGroup{}
	for _, c := range clients {
		wg.Add(1)
		go func(client xproviders.Uploader) {
			if uploadErr := client.Upload(context.Background(), bucket, key, file); uploadErr != nil {
				uploadErrors <- uploadErr
			}
			wg.Done()
		}(c)
	}

	wg.Wait()
	close(uploadErrors)
	if len(uploadErrors) > 0 {
		fmt.Printf("%d Errors occured during upload\n:", len(uploadErrors))
	}
	for e := range uploadErrors {
		fmt.Println(e)
	}
}

func validateProviders(providers []Provider) error {
	if len(providers) == 0 {
		return errors.New("providers cannot be empty")
	}
	b := strings.Builder{}
	for _, p := range providers {
		if _, ok := ValidProviders[p]; !ok {
			b.WriteString(fmt.Sprintf("%q is not a valid provider\n", p))
		}
	}
	if b.Len() != 0 {
		return errors.New(b.String())
	} else {
		return nil
	}
}

func initProvider(ctx context.Context, p Provider, cfg *config.Config) (xproviders.Uploader, error) {
	switch p {
	case AWS:
		return initAWS(cfg.GetAWS())
	case GCP:
		return initGCP(ctx, cfg.GetGCP())
	case Azure:
		return initAzure(cfg.GetAzure())
	default:
		return nil, errors.New("unknown provider")
	}
}

func initAWS(cfg config.AWSConfig) (*aws.AWSUploader, error) {
	if cfg.Credentials.Profile == "" || cfg.Credentials.Filename == "" {
		return nil, errors.New("AWS config invalid")
	}

	client, err := aws.New(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initGCP(ctx context.Context, cfg config.GCPConfig) (*gcp.GCPUploader, error) {
	if cfg.Credentials.Filename == "" || len(cfg.Credentials.Scopes) == 0 {
		return nil, errors.New("invalid gcp config")
	}
	client, err := gcp.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initAzure(cfg config.AzureConfig) (*azure.AzureUploader, error) {
	if cfg.Credentials.AccountName == "" || cfg.Credentials.AccountKey == "" {
		return nil, errors.New("invalid azure config")
	}
	client, err := azure.New(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
