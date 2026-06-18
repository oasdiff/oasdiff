package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestValueDiff_Reverse(t *testing.T) {
	d := &diff.ValueDiff{
		From: "a",
		To:   "b",
	}

	require.Equal(t, &diff.ValueDiff{
		From: "b",
		To:   "a",
	}, d.Reverse())
}

func TestValueDiff_ReverseNil(t *testing.T) {
	var d *diff.ValueDiff
	require.Nil(t, d.Reverse())
}
