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
		"origin/-upload-pack=/bin/sh",
		"remotes/origin/-x",
		"refs/remotes/origin/-output=/tmp/pwned",
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
		"origin/feature/dash-in-name",
	} {
		t.Run(ref, func(t *testing.T) {
			require.NoError(t, checkRef(ref))
		})
	}
}

func TestParseOriginBranch(t *testing.T) {
	tests := []struct {
		ref        string
		wantBranch string
		wantOK     bool
	}{
		{"origin/main", "main", true},
		{"origin/feature/login", "feature/login", true},
		{"remotes/origin/main", "main", true},
		{"refs/remotes/origin/main", "main", true},
		{"refs/remotes/origin/release/1.0", "release/1.0", true},
		{"main", "", false},
		{"HEAD", "", false},
		{"b09dd8d44cc05f88ad3364434f662e2e2e33db37", "", false},
		{"origin/", "", false},
		{"origin/ ", "", false},
		{"origin/branch:withcolon", "", false},
		{"refs/heads/main", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.ref, func(t *testing.T) {
			branch, ok := parseOriginBranch(tc.ref)
			require.Equal(t, tc.wantOK, ok)
			require.Equal(t, tc.wantBranch, branch)
		})
	}
}
