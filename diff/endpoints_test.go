package diff_test

import (
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestEndpointsSort(t *testing.T) {
	endpoints := diff.Endpoints{
		{
			Method: "GET",
			Path:   "/b",
		},
		{
			Method: "GET",
			Path:   "/a",
		},
	}

	slices.SortFunc(endpoints, endpoints.SortFunc)
	require.Equal(t, "/a", endpoints[0].Path)
}

func TestEndpointsSort_Methods(t *testing.T) {
	endpoints := diff.Endpoints{
		{
			Method: "POST",
			Path:   "/a",
		},
		{
			Method: "OPTIONS",
			Path:   "/a",
		},
	}

	slices.SortFunc(endpoints, endpoints.SortFunc)
	require.Equal(t, "OPTIONS", endpoints[0].Method)
}
