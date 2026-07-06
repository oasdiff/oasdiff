package checker

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ComputeFingerprint produces a short, stable identifier for a change or
// finding. It is used to match the same logical item across spec versions
// (carry-forward of review state in oasdiff-service) and to look up review
// records at view time. The changelog and validate commands share it so a
// downstream tool can match findings produced by either.
//
// Inputs are the structured rule arguments rather than the rendered text:
// rendered text varies with locale and copy edits to the message templates,
// which would silently invalidate every stored fingerprint. The args carry the
// same per-change disambiguation power without that fragility.
func ComputeFingerprint(id, operation, path string, args []any) string {
	h := fmt.Sprintf("%s:%s:%s:%s", id, operation, path, serializeArgs(args))
	sum := sha256.Sum256([]byte(h))
	return hex.EncodeToString(sum[:])[:12]
}

// serializeArgs joins the change args into a deterministic string. fmt's `%v`
// verb sorts map keys (Go 1.12+) so nested maps remain stable; primitives and
// slices are stable by definition.
func serializeArgs(args []any) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = fmt.Sprintf("%v", a)
	}
	return strings.Join(parts, ";")
}
