package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestStringsDiff_Reverse(t *testing.T) {
	d := &diff.StringsDiff{
		Added:   []string{"a"},
		Deleted: []string{"b", "c"},
	}

	require.Equal(t, &diff.StringsDiff{
		Added:   []string{"b", "c"},
		Deleted: []string{"a"},
	}, d.Reverse())
}

func TestStringsDiff_ReverseNil(t *testing.T) {
	var d *diff.StringsDiff
	require.Nil(t, d.Reverse())
}
