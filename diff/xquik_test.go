package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestXquikOpenAPI31TweetSchemaPropertyAdded(t *testing.T) {
	s1 := loadSpec(t, "../data/xquik/base.yaml")
	s2 := loadSpec(t, "../data/xquik/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.ComponentsDiff)
	require.NotNil(t, d.ComponentsDiff.SchemasDiff)

	tweetDiff := d.ComponentsDiff.SchemasDiff.Modified["Tweet"]
	require.NotNil(t, tweetDiff)
	require.NotNil(t, tweetDiff.PropertiesDiff)
	require.Contains(t, tweetDiff.PropertiesDiff.Added, "lang")
}
