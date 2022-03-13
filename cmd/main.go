package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/stevequadros/uploader/config"
	xproviders "github.com/stevequadros/uploader/providers"
	"github.com/stevequadros/uploader/providers/coordinator"
	"os"
	"strings"
)

// verify file was uploaded code for ease of use - maybe
// configs need proper json annotations
// ease of testing
//	- move all config validation into config
//  - move all other code to providers.Coordinator

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

var usage = `
uploader uploads a file to any of the provider providers [aws, gcp, azure] to the given bucket, key.

see example_config.json to get started on your config file. 

Example Usage:
./uploader --provider aws --provider azure --provider gcp --file test.txt --config ~/.filescom/config.json -bucket filescometestagain -key test.txt
`

func main() {
	providers := providerFlag{}
	var filename, configPath, bucket, key string

	if err := validateFlags(&providers, &filename, &configPath, &bucket, &key); err != nil {
		logErrorAndExit("Error processing flags", err)
	}

	logInProcess("Validating config")
	cfg, err := config.New(configPath)
	if err != nil {
		logErrorAndExit("Config error", err)
	}
	logSuccess("Config validated")

	logInProcess("Checking File to upload...")
	file := validateUploadFile(filename)
	defer func() {
		if err = file.Close(); err != nil {
			logErrorAndExit("error closing upload file", err)
		}
	}()
	logSuccess("Upload file valid")

	logInProcess("Initializing Providers...")
	ctx := context.Background()
	var coord coordinator.Coordinator
	coord, err = coordinator.NewCoordinator(ctx, cfg)
	if err != nil {
		logErrorAndExit("Error intializing providers", err)
	}
	logSuccess(fmt.Sprintf("Providers Initialized: %v", providers))

	logInProcess("Beginning Uploads")
	res, err := coord.Do(ctx, bucket, key, file)
	if err != nil {
		logError("error uploading", err)
	}
	for _, p := range res.Done {
		logSuccess(fmt.Sprintf("Successfully Uploaded to %q", p))
	}

	for _, e := range res.Failed {
		logError(fmt.Sprintf("Error Uploading file to %q: ", e.Provider), e.Error)
	}
	logSuccess(fmt.Sprintf("file %q uploaded to %d providers: %v", file.Name(), len(coord.Providers()), coord.Providers()))
	os.Exit(0)
}

func validateFlags(providers *providerFlag, filename, configPath, bucket, key *string) error {
	flag.Var(providers, "provider", "[REQUIRED 1+] Providers targeted. Valid Options: aws, gcp, azure. Each one must be preceded with it's own flag, ex: -provider aws -provider azure -provider gcp")
	flag.StringVar(filename, "file", "", "[REQUIRED] The file to upload")
	flag.StringVar(configPath, "config", "", "[REQUIRED] Path to config.json")
	flag.StringVar(bucket, "bucket", "", "[REQUIRED] Target bucket for file. Will Create bucket if it doesn't exist.")
	flag.StringVar(key, "key", "", "[REQUIRED] key for file")
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
	if err := validateProviders(*providers); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if *configPath == "" {
		validationErrors = append(validationErrors, errors.New("configPath flag cannot be empty"))
	}

	if *filename == "" {
		validationErrors = append(validationErrors, errors.New("filename flag to upload cannot be empty"))
	}

	if *bucket == "" {
		validationErrors = append(validationErrors, errors.New("bucket flag cannot be empty"))
	}

	if *key == "" {
		validationErrors = append(validationErrors, errors.New("key flag cannot be empty"))
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

func validateUploadFile(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		handleErrAndExit("could not open file to upload ", err)
	}
	return file
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

func logErrorAndExit(prepend string, e error) {
	logError(prepend, e)
	os.Exit(1)
}

func logSuccess(s string) {
	fmt.Println("✓ " + s)
}
