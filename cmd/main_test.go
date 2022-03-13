package main

import (
	"fmt"
	xproviders "github.com/stevequadros/uploader/providers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidProviders(t *testing.T) {
	tc := map[string]struct {
		providers []xproviders.Provider
		err       bool
	}{
		"valid providers has no error":  {[]xproviders.Provider{xproviders.GCP, xproviders.AWS, xproviders.Azure}, false},
		"no providers is an error":      {[]xproviders.Provider{}, true},
		"invalid providers is an error": {[]xproviders.Provider{xproviders.GCP, xproviders.AWS, xproviders.Azure, "Foo"}, true},
	}

	for _, tt := range tc {
		t.Run(fmt.Sprintf("%v", tt.providers), func(t *testing.T) {
			actual := validateProviders(tt.providers)
			if tt.err {
				require.Error(t, actual)
			} else {
				require.NoError(t, actual)
			}
		})
	}
}

func Test_validateFlags(t *testing.T) {
	type args struct {
		providers  providerFlag
		filename   string
		configPath string
		bucket     string
		key        string
	}

	filename, configPath, bucket, key := "test", "test", "test", "test"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid flags no errors", args{providerFlag{"aws"}, filename, configPath, bucket, key}, false},
		{"filename blank errors", args{providerFlag{"aws"}, "", configPath, bucket, key}, true},
		{"configpath blank errors", args{providerFlag{"aws"}, filename, "", bucket, key}, true},
		{"bucket blank errors", args{providerFlag{"aws"}, filename, configPath, "", key}, true},
		{"key blank errors", args{providerFlag{"aws"}, filename, configPath, bucket, ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateFlags(tt.args.providers, tt.args.filename, tt.args.configPath, tt.args.bucket, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("validateFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
