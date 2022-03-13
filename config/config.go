package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Config struct {
	AWS   *AWS
	Azure *Azure
	GCP   *GCP
}

/*
Config takes list of providers and path to config file
validates config file
validates each provider config
*/

func New(path string) (config Config, err error) {
	var configFile *os.File
	configFile, err = os.Open(path)
	defer func() {
		err = configFile.Close()
	}()
	config, err = NewFromJSON(configFile)
	return config, err
}

func NewFromJSON(reader io.Reader) (config Config, err error) {
	var configData []byte
	configData, err = io.ReadAll(reader)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		return config, err
	}

	err = config.Validate()
	if err != nil {
		return config, err
	}
	return config, nil
}

func (c *Config) Validate() error {
	if c.AWS != nil {
		if err := c.AWS.Validate(); err != nil {
			return err
		}
	}
	if c.Azure != nil {
		if err := c.Azure.Validate(); err != nil {
			return err
		}
	}
	if c.GCP != nil {
		if err := c.GCP.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type AWS struct {
	Credentials *AWSCredentials
}

type AWSCredentials struct {
	// location of aws credentials file
	Filename string
	// profile to use
	Profile string
}

func NewAWS(filename, profile string) *AWS {
	return &AWS{
		Credentials: &AWSCredentials{
			Filename: filename,
			Profile:  profile,
		},
	}
}

func (p *AWS) Validate() error {
	if p.Credentials == nil {
		return errors.New("empty credentials")
	}
	if p.Credentials.Profile == "" {
		return errors.New("aws profile empty")
	}
	if p.Credentials.Filename == "" {
		return errors.New("aws config filename empty")
	}
	return nil
}

type Azure struct {
	Credentials *AzureCredentials
}

type AzureCredentials struct {
	AccountName string
	AccountKey  string
}

func NewAzure(accountName, accountKey string) *Azure {
	return &Azure{Credentials: &AzureCredentials{
		AccountName: accountName,
		AccountKey:  accountKey,
	}}
}

func (p *Azure) Validate() error {
	if p.Credentials == nil {
		return errors.New("empty credentials")
	}
	if p.Credentials.AccountName == "" {
		return errors.New("azure accountname empty")
	}
	if p.Credentials.AccountKey == "" {
		return errors.New("azure key empty")
	}
	return nil
}

type GCP struct {
	Credentials *GCPCredentials
}

type GCPCredentials struct {
	// path to json GCP credentials
	Filename string
	Scopes   []string
}

func NewGCP(filename string) *GCP {
	return &GCP{&GCPCredentials{
		Filename: filename,
		// defaulting to this for now, ideally should accept scopes as argument in more fleshed out version
		Scopes: []string{"https://www.googleapis.com/auth/devstorage.full_control"},
	},
	}
}

func (p *GCP) Validate() error {
	if p.Credentials == nil {
		return errors.New("empty credentials")
	}
	if p.Credentials.Filename == "" {
		return errors.New("gcp filename empty")
	}
	if len(p.Credentials.Scopes) == 0 {
		return errors.New("gcp scopes empty")
	}
	return nil
}
