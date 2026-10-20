package diff

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMapDiff_KeysSorted(t *testing.T) {
	base := map[string]string{"common": "v"}
	revision := map[string]string{"common": "v"}
	var keys []string
	for i := range 20 {
		key := fmt.Sprintf("k%02d", i)
		keys = append(keys, key)
		base["deleted-"+key] = "v"
		revision["added-"+key] = "v"
	}

	d := getStringMapDiff(base, revision)
	for i, key := range keys {
		require.Equal(t, "added-"+key, d.Added[i])
		require.Equal(t, "deleted-"+key, d.Deleted[i])
	}
}
