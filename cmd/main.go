package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers/aws"
	"github.com/stevequadros/uploader/providers/azure"
	"github.com/stevequadros/uploader/providers/gcp"
	"log"
	"os"
	"strings"
)

// providers
// autocreate bucket default to false, prompt if bucket not exists
// Gcp complains if you try to create an existing bucket TOP PRIORITY
// - move bucket creation to interface function for easy testing of paths
// - move uploader behind interface for easy testing`
// verify file was uploaded code for ease of use - maybe
// uploading files should attempt all and return list of errors rather than fast failing
// os.Getenv for additional option to config file

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
	var filename, configPath string
	flag.Var(&providers, "provider", "Providers targeted. Valid Options: aws, gcp, azure. Each one must be preceded with it's own flag, ex: -provider aws -provider azure -provider gcp")
	flag.StringVar(&filename, "file", "", "The file to upload")
	flag.StringVar(&configPath, "config", "", "Path to config.json")
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

	ctx := context.Background()
	var uploadErrors []error
	for _, p := range providers {
		if err = initProvider(p, ctx, cfg, file); err != nil {
			uploadErrors = append(uploadErrors, err)
		}
	}
	if len(uploadErrors) > 0 {
		for _, e := range uploadErrors {
			fmt.Println(e)
		}
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

func initProvider(p Provider, ctx context.Context, cfg *config.Config, file *os.File) error {
	switch p {
	case AWS:
		return initAWS(ctx, cfg.AWS, file)
	case GCP:
		return initGCP(ctx, cfg.GCP, file)
	case Azure:
		return initAzure(ctx, cfg.Azure, file)
	default:
		return errors.New("unknown provider")
	}
}

func initAWS(ctx context.Context, cfg *config.AWSConfig, file *os.File) error {
	if cfg == nil || cfg.Credentials.Profile == "" || cfg.Credentials.Filename == "" {
		return errors.New("AWS config invalid")
	}

	awsClient, err := aws.New(cfg)
	if err != nil {
		return err
	}
	if err = awsClient.Upload(ctx, fmt.Sprintf("%s-%s", "filescom-takehome-", "test"), file.Name(), file); err != nil {
		return err
	}
	return nil
}

func initGCP(ctx context.Context, cfg *config.GCPConfig, file *os.File) error {
	if cfg == nil || cfg.Credentials.Filename == "" || len(cfg.Credentials.Scopes) == 0 {
		return errors.New("invalid gcp config")
	}
	client, err := gcp.New(ctx, cfg)
	if err != nil {
		return err
	}
	if err = client.Upload(ctx, fmt.Sprintf("%s-%s", "filescom-takehome-", "test"), file.Name(), file); err != nil {
		return err
	}
	return nil
}

func initAzure(ctx context.Context, cfg *config.AzureConfig, file *os.File) error {
	if cfg == nil || cfg.Credentials.AccountName == "" || cfg.Credentials.AccountKey == "" {
		return errors.New("invalid azure config")
	}
	azureClient, err := azure.New(cfg)
	if err != nil {
		log.Fatal("error creating azure", err)
	}
	if err = azureClient.Upload(ctx, "filescom-takehome", file.Name(), file); err != nil {
		log.Fatal("error uploading azure", err)
	}
	return nil
}
