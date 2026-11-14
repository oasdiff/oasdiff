package build

import "testing"

func TestVersion(t *testing.T) {
	for _, tt := range []struct {
		name     string
		injected string
		module   string
		want     string
	}{
		{"a release build injects the version", "1.38.0", "", "1.38.0"},
		{"an injected version wins over the module", "1.38.0", "v1.37.0", "1.38.0"},
		{"go install reports the module version, without its v", unset, "v1.38.0", "1.38.0"},
		{"a working-tree build reports the tag it is at", unset, "v1.38.0", "1.38.0"},
		{"an edited tree says so", unset, "v1.38.0+dirty", "1.38.0+dirty"},
		{"a commit past the tag reports the pseudo-version", unset, "v1.38.1-0.20261114144223-679ecb174da7", "1.38.1-0.20261114144223-679ecb174da7"},
		{"a build outside a checkout has no version to report", unset, "(devel)", unset},
		{"a binary with no build info has none either", unset, "", unset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := version(tt.injected, tt.module); got != tt.want {
				t.Errorf("version(%q, %q) = %q, want %q", tt.injected, tt.module, got, tt.want)
			}
		})
	}
}
