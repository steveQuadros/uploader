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
	"os"
	"strings"
	"sync"
)

// providers
// verify file was uploaded code for ease of use - maybe
// uploading files should attempt all and return list of errors rather than fast failing
// os.Getenv for additional option to config file
// configs need proper json annotations

type ProviderError struct {
	p   xproviders.Provider
	err error
}

type providerFlag []xproviders.Provider

func (i *providerFlag) String() string {
	b := strings.Builder{}
	for _, p := range *i {
		b.WriteString(string(p))
	}
	return b.String()
}

func (i *providerFlag) Set(value string) error {
	*i = append(*i, xproviders.Provider(value))
	return nil
}

var providers providerFlag
var filename, configPath, bucket, key string

var usage = `
uploader uploads a file to any of the provider providers [aws, gcp, azure] to the given bucket, key

Example Usage:
./uploader --provider aws --provider azure --provider gcp --file test.txt --config ~/.filescom/config.json -bucket filescometestagain -key test.txt
`

func main() {
	handleErrAndExit("Error validating flags", validateFlags())

	logInProcess("Validating config")
	configFile, err := os.Open(configPath)
	defer func() {
		handleErrAndExit("Error closing config file: ", configFile.Close())
	}()
	if err != nil {
		handleErrAndExit("error opening config ", err)
	}
	cfg, err := config.NewFromJSON(configFile)
	if err != nil {
		handleErrAndExit("error parsing configfile ", err)
	}
	logSuccess("Config validated")

	logInProcess("Checking File to upload...")
	file, err := os.Open(filename)
	defer func() {
		if err = file.Close(); err != nil {
			handleErrAndExit("error closing upload file", err)
		}
	}()
	if err != nil {
		handleErrAndExit("could not open file to upload ", err)
	}
	logSuccess("Upload file valid")

	ctx := context.Background()

	logInProcess("Initializing Providers...")
	var clients []xproviders.Uploader
	var initErrors []error
	for _, p := range providers {
		var client xproviders.Uploader
		client, err = initProvider(ctx, p, cfg)
		if err != nil {
			initErrors = append(initErrors, err)
		}
		clients = append(clients, client)
	}

	if len(initErrors) != 0 {
		fmt.Println("\u2717 Error(s) initializing Providers")
		for _, e := range initErrors {
			fmt.Println(e)
		}
		os.Exit(1)
	}
	logSuccess(fmt.Sprintf("Providers Initialized: %v", providers))

	logInProcess("Beginning Uploads")
	uploadErrors := make(chan ProviderError, len(providers))
	success := make(chan xproviders.Provider, len(providers))
	wg := sync.WaitGroup{}
	var count int
	for i, c := range clients {
		wg.Add(1)
		go func(client xproviders.Uploader, n int) {
			p := providers[n]
			if uploadErr := client.Upload(context.Background(), bucket, key, file); uploadErr != nil {
				uploadErrors <- ProviderError{p, uploadErr}
			} else {
				success <- p
			}
			wg.Done()
		}(c, i)
	}

	for count < len(providers) {
		select {
		case e := <-uploadErrors:
			logError(fmt.Sprintf("Error Uploading file to %q: ", e.p), e.err)
			count++
		case p := <-success:
			logSuccess(fmt.Sprintf("Successfully Uploaded to %q", p))
			count++
		}
	}

	wg.Wait()
	close(uploadErrors)
	close(success)
	os.Exit(0)
}

func validateFlags() error {
	flag.Var(&providers, "provider", "[REQUIRED 1+] Providers targeted. Valid Options: aws, gcp, azure. Each one must be preceded with it's own flag, ex: -provider aws -provider azure -provider gcp")
	flag.StringVar(&filename, "file", "", "[REQUIRED] The file to upload")
	flag.StringVar(&configPath, "config", "", "[REQUIRED] Path to config.json")
	flag.StringVar(&bucket, "bucket", "", "[REQUIRED] Target bucket for file. Will Create bucket if it doesn't exist.")
	flag.StringVar(&key, "key", "", "[REQUIRED] key for file")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage = func() {
			fmt.Println(usage)
			flag.PrintDefaults()
		}
		flag.Usage()
		os.Exit(1)
	}

	var validationErrors []error
	if err := validateProviders(providers); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if configPath == "" {
		validationErrors = append(validationErrors, errors.New("configPath cannot be empty"))
	}

	if filename == "" {
		validationErrors = append(validationErrors, errors.New("filename to upload cannot be empty"))
	}

	if bucket == "" {
		validationErrors = append(validationErrors, errors.New("bucket cannot be empty"))
	}

	if key == "" {
		validationErrors = append(validationErrors, errors.New("key cannot be empty"))
	}

	if len(validationErrors) > 0 {
		b := strings.Builder{}
		for _, e := range validationErrors {
			b.WriteString(e.Error() + "\n")
		}
		return errors.New(b.String())
	} else {
		return nil
	}
}

func validateProviders(providers []xproviders.Provider) error {
	if len(providers) == 0 {
		return errors.New("providers cannot be empty")
	}
	b := strings.Builder{}
	for _, p := range providers {
		if _, ok := xproviders.Providers[p]; !ok {
			b.WriteString(fmt.Sprintf("%q is not a valid provider\n", p))
		}
	}
	if b.Len() != 0 {
		return errors.New(b.String())
	} else {
		return nil
	}
}

func initProvider(ctx context.Context, p xproviders.Provider, cfg *config.Config) (xproviders.Uploader, error) {
	switch p {
	case xproviders.AWS:
		return initAWS(cfg.GetAWS())
	case xproviders.GCP:
		return initGCP(ctx, cfg.GetGCP())
	case xproviders.Azure:
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

func handleErrAndExit(prepend string, err error) {
	if err != nil {
		logError(prepend, err)
		os.Exit(1)
	}
}

func logInProcess(s string) {
	fmt.Println(s + "...")
}

func logError(prepend string, e error) {
	fmt.Println("✗ " + prepend)
	fmt.Println(e)
}

func logSuccess(s string) {
	fmt.Println("✓ " + s)
}
