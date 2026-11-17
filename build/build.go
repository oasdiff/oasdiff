package build

import (
	"runtime/debug"
	"strings"
)

// unset is what Version holds in a build that does not inject one.
const unset = "main"

// Version of the binary, to be injected at build time.
var Version = unset

// String is the version to report. A release build injects it; a binary from
// "go install <module>@<version>" carries the version in its build info
// instead, which is the only place it exists there.
func String() string {
	return version(Version, moduleVersion())
}

// version prefers the injected version and otherwise reports the module
// version the binary was built from: the tag for "go install <module>@<tag>"
// and for a clean checkout of it, that tag plus "+dirty" for an edited tree,
// and a pseudo-version for a commit past it. A build outside a checkout
// records "(devel)", which names nothing, so that keeps the default.
func version(injected, module string) string {
	if injected != unset {
		return injected
	}
	if module == "" || module == "(devel)" {
		return injected
	}
	// A release injects the tag without its "v", so both paths read alike.
	return strings.TrimPrefix(module, "v")
}

func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return info.Main.Version
}
