package load

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Refs reach git as operands, so a leading dash is parsed as an OPTION rather
// than a revision. That is a real capability, not a theoretical one:
// "git show --output=<path> <rev>" writes to that path (arbitrary file
// overwrite as the invoking user), and "git fetch --upload-pack=<program>"
// makes git execute the program. Verified against git 2.49.
//
// Call sites pass --end-of-options; checkRef is the version-independent belt to
// that braces, and gives a clear error instead of a confusing git one.
func TestCheckRefRejectsOptionLikeRevisions(t *testing.T) {
	for _, ref := range []string{
		"--output=/tmp/pwned",
		"--upload-pack=/bin/sh",
		"-x",
		"--end-of-options",
	} {
		t.Run(ref, func(t *testing.T) {
			require.Error(t, checkRef(ref), "a ref git would parse as an option must be rejected")
		})
	}
}

func TestCheckRefAllowsRealRevisions(t *testing.T) {
	for _, ref := range []string{
		"HEAD",
		"main",
		"origin/main",
		"v1.2.3",
		"b09dd8d44cc05f88ad3364434f662e2e2e33db37",
		"feature/dash-in-name",
	} {
		t.Run(ref, func(t *testing.T) {
			require.NoError(t, checkRef(ref))
		})
	}
}
