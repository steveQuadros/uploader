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
