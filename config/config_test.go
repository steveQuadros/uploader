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
      "filename": "gcpfilename",
      "scopes": ["scope1", "scope2"]
    }
  }
}
`

func TestNewFromJSONParsing(t *testing.T) {
	b := bytes.NewBuffer([]byte(validConfig))
	config, err := NewFromJSON(b)
	require.NoError(t, err)
	require.Equal(t, "/.aws/credentials", config.AWS.Credentials.Filename)
	require.Equal(t, "testprofile", config.AWS.Credentials.Profile)
	require.Equal(t, "azureaccountname", config.Azure.Credentials.AccountName)
	require.Equal(t, "azurekey", config.Azure.Credentials.AccountKey)
	require.Equal(t, "gcpfilename", config.GCP.Credentials.Filename)
}

func TestNewFromJSON(t *testing.T) {
	tc := map[string]struct {
		in       string
		expected Config
		err      bool
	}{
		"valid config": {
			validConfig,
			Config{
				AWS: &AWS{
					Credentials: &AWSCredentials{Filename: "/.aws/credentials", Profile: "testprofile"},
				},
				Azure: &Azure{Credentials: &AzureCredentials{AccountName: "azureaccountname", AccountKey: "azurekey"}},
				GCP: &GCP{Credentials: &GCPCredentials{
					Filename: "gcpfilename",
					Scopes:   []string{"scope1", "scope2"},
				}},
			},
			false,
		},
		"invalid json returns an error": {
			`{"aws":""""}`,
			Config{},
			true,
		},
		"[aws] config invalid without filename": {
			`{"aws":{}}`,
			Config{AWS: &AWS{}},
			true,
		},
		"[aws] config invalid without profile": {
			`{"aws": {"credentials": {"filename": "test"}}}`,
			Config{AWS: &AWS{&AWSCredentials{Filename: "test"}}},
			true,
		},
		"[gcp] config invalid without filename": {
			`{"gcp":{}}`,
			Config{GCP: &GCP{}},
			true,
		},
		"[gcp] config invalid without scopes": {
			`{"gcp": {"credentials": {"filename": "test"}}}`,
			Config{GCP: &GCP{&GCPCredentials{Filename: "test"}}},
			true,
		},
		"[azure] config invalid without accountname": {
			`{"azure":{}}`,
			Config{Azure: &Azure{}},
			true,
		},
		"[azure] config invalid without account key": {
			`{"azure": {"credentials": {"accountName": "test"}}}`,
			Config{Azure: &Azure{&AzureCredentials{AccountName: "test"}}},
			true,
		},
	}

	for name, tt := range tc {
		t.Run(name, func(t *testing.T) {
			cfg, err := NewFromJSON(bytes.NewReader([]byte(tt.in)))
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, cfg)
		})
	}
}
