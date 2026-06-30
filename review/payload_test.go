package review

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestManifest(t *testing.T) {
	changes := checker.Changes{
		checker.ApiChange{Id: "id-a", Operation: "GET", Path: "/a", Level: checker.ERR},
		checker.ApiChange{Id: "id-b", Operation: "POST", Path: "/b", Level: checker.INFO},
	}
	manifest := Manifest(changes)
	require.Len(t, manifest, 2)
	for _, m := range manifest {
		require.NotEmpty(t, m.Fingerprint, "every entry must carry a fingerprint")
		require.Len(t, m.Fingerprint, 12, "fingerprint is the 12-char ComputeFingerprint output")
	}
	require.Equal(t, int(checker.ERR), manifest[0].Level, "level is the change's severity as an int")
	require.Equal(t, int(checker.INFO), manifest[1].Level)
	require.NotEqual(t, manifest[0].Fingerprint, manifest[1].Fingerprint, "distinct changes must have distinct fingerprints")
}
