package review

import (
	"encoding/json"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// Payload is the cleartext review bundle. It is AES-256-GCM-encrypted with a
// fresh key (see Encrypt) and only the ciphertext is uploaded, so the server
// receives a blob it cannot read; a decryptor on the rendering side reconstructs
// it. This package is the single source of truth for the bundle's wire shape --
// that decryptor mirrors these json tags.
//
// BaseSpec / RevisionSpec hold each spec's bytes verbatim as a string. A YAML
// spec stays YAML text and a JSON spec stays JSON text -- this JSON object is
// only the envelope that bundles the several fields into one blob; it does not
// reformat or reparse the spec content.
//
// Changes is the JSON changelog the caller already computed, embedded raw. The
// server can't recompute it (it can't read the specs), so it ships in the
// bundle and the renderer consumes it directly.
//
// Blocks is the per-change structural slices the UI renders instead of the full
// spec (see Extract). Empty when extraction finds nothing sliceable, in which
// case the renderer falls back to the full spec.
type Payload struct {
	BaseSpec         string          `json:"base_spec" yaml:"base_spec"`
	RevisionSpec     string          `json:"revision_spec" yaml:"revision_spec"`
	BaseFilename     string          `json:"base_filename" yaml:"base_filename"`
	RevisionFilename string          `json:"revision_filename" yaml:"revision_filename"`
	Changes          json.RawMessage `json:"changes" yaml:"changes"`
	Mode             string          `json:"mode" yaml:"mode"`
	Blocks           []Block         `json:"blocks,omitempty" yaml:"blocks,omitempty"`
}

// Change is one manifest entry sent alongside the encrypted bundle on the
// authenticated (Pro) path: a change's fingerprint (see
// formatters.ComputeFingerprint) and its level.
type Change struct {
	Fingerprint string `json:"fingerprint" yaml:"fingerprint"`
	Level       int    `json:"level" yaml:"level"`
}

// Manifest builds the {fingerprint, level} manifest. Fingerprints are computed
// as the JSON formatter does, so they match the ones in the encrypted changelog
// the page renders and the fingerprints carried on each Block.
func Manifest(changes checker.Changes) []Change {
	manifest := make([]Change, 0, len(changes))
	for _, change := range changes {
		manifest = append(manifest, Change{
			Fingerprint: formatters.ComputeFingerprint(change.GetId(), change.GetOperation(), change.GetPath(), change.GetArgs()),
			Level:       int(change.GetLevel()),
		})
	}
	return manifest
}
