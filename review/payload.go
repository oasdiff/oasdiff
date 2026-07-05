package review

import (
	"encoding/json"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// Payload is the cleartext review bundle and the single source of truth for
// its wire shape: a decryptor on the receiving side mirrors these json tags.
// Encrypt seals it with a fresh key and only the ciphertext is uploaded, so
// the server receives a blob it cannot read.
//
// BaseSpec/RevisionSpec hold each spec's bytes verbatim (YAML stays YAML text);
// this JSON object is only the envelope. Changes is the changelog the caller
// already computed, embedded raw: the server can't recompute what it can't
// read. Blocks is the per-change structural slices (see Extract); empty when
// the changelog is empty, since every change resolves to some block.
type Payload struct {
	BaseSpec         string          `json:"base_spec" yaml:"base_spec"`
	RevisionSpec     string          `json:"revision_spec" yaml:"revision_spec"`
	BaseFilename     string          `json:"base_filename" yaml:"base_filename"`
	RevisionFilename string          `json:"revision_filename" yaml:"revision_filename"`
	Changes          json.RawMessage `json:"changes" yaml:"changes"`
	Mode             string          `json:"mode" yaml:"mode"`
	// Composed marks a bundle built from a set of specs per side (composed
	// mode): there is no single spec or filename, so BaseSpec/RevisionSpec and
	// the filenames are empty and the blocks carry the comparison.
	Composed bool    `json:"composed,omitempty" yaml:"composed,omitempty"`
	Blocks   []Block `json:"blocks,omitempty" yaml:"blocks,omitempty"`
}

// Change is one manifest entry sent alongside the encrypted bundle in
// cleartext: a change's fingerprint (see formatters.ComputeFingerprint) and
// its level, so a server can track per-change state without reading the bundle.
type Change struct {
	Fingerprint string `json:"fingerprint" yaml:"fingerprint"`
	Level       int    `json:"level" yaml:"level"`
}

// Manifest builds the {fingerprint, level} manifest. Fingerprints are computed
// as the JSON formatter does, so they match the ones in the encrypted changelog
// the bundle carries and the fingerprints on each Block.
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
