package config

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

var validConfig = `
{
  "aws": {
    "credentials": {
      "filename": "/.aws/credentials",
      "profile": "testprofile"
    }
  },
  "azure": {
    "credentials": {
      "accountName": "azureaccountname",
      "accountKey": "azurekey"
    }
  },
  "gcp": {
    "credentials": {
      "filename": "gcpfilename"
    }
  }
}
`

func TestNewFromJSON(t *testing.T) {
	b := bytes.NewBuffer([]byte(validConfig))
	config, err := NewFromJSON(b)
	require.NoError(t, err)
	require.Equal(t, "/.aws/credentials", config.AWS.Credentials.Filename)
	require.Equal(t, "testprofile", config.AWS.Credentials.Profile)
	require.Equal(t, "azureaccountname", config.Azure.Credentials.AccountName)
	require.Equal(t, "azurekey", config.Azure.Credentials.AccountKey)
	require.Equal(t, "gcpfilename", config.GCP.Credentials.Filename)
}
