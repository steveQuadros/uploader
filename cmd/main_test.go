package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidProviders(t *testing.T) {
	tc := map[string]struct {
		providers []Provider
		err       bool
	}{
		"valid providers has no error":  {[]Provider{GCP, AWS, Azure}, false},
		"no providers is an error":      {[]Provider{}, true},
		"invalid providers is an error": {[]Provider{GCP, AWS, Azure, "Foo"}, true},
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
