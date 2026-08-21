package utils_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/utils"
	"github.com/stretchr/testify/require"
)

func TestToKebabCase(t *testing.T) {
	for in, want := range map[string]string{
		"maxItems":         "max-items",
		"exclusiveMaximum": "exclusive-maximum",
		"openIdConnectUrl": "open-id-connect-url",
		"type":             "type",
		"":                 "",
		"Ref":              "ref",
	} {
		require.Equal(t, want, utils.ToKebabCase(in), in)
	}
}
