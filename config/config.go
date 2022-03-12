package config

import (
	"encoding/json"
	"io"
)

type Config struct {
	AWS   *AWSConfig
	Azure *AzureConfig
	GCP   *GCPConfig
}

func NewFromJSON(reader io.Reader) (*Config, error) {
	var configData []byte
	configData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// @TODO make unexported unnecessary types

type AWSConfig struct {
	Credentials *AWSCredentials
}

type AWSCredentials struct {
	// location of aws credentials file
	Filename string
	// profile to use
	Profile string
}

func NewAWS(filename, profile string) *AWSConfig {
	return &AWSConfig{
		Credentials: &AWSCredentials{
			Filename: filename,
			Profile:  profile,
		},
	}
}

type AzureConfig struct {
	Credentials *AzureCredentials
}

type AzureCredentials struct {
	AccountName string
	AccountKey  string
}

func NewAzure(accountName, accountKey string) *AzureConfig {
	return &AzureConfig{Credentials: &AzureCredentials{
		AccountName: accountName,
		AccountKey:  accountKey,
	}}
}

type GCPConfig struct {
	Credentials *GCPCredentials
}

type GCPCredentials struct {
	// path to json GCP credentials
	Filename string
	Scopes   []string
}

func NewGCP(filename string) *GCPConfig {
	return &GCPConfig{&GCPCredentials{
		Filename: filename,
		// defaulting to this for now, ideally should accept scopes as argument in more fleshed out version
		Scopes: []string{"https://www.googleapis.com/auth/devstorage.full_control"},
	},
	}
}
