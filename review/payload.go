package review

import (
	"encoding/json"

	"github.com/oasdiff/oasdiff/checker"
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
	// ToolVersion is the oasdiff version that produced the bundle (build.Version,
	// e.g. "v1.25.1"; "main" for a dev build). It travels inside the encrypted
	// bundle, so the receiving review page can tell when a review came from an
	// outdated oasdiff and nudge an upgrade. Empty for bundles from older
	// clients that predate this field. The GitHub Action inherits it, it runs
	// this same CLI with --open.
	ToolVersion string `json:"tool_version,omitempty" yaml:"tool_version,omitempty"`
	// Platform is where the bundle was produced, from the PLATFORM environment
	// variable: "github-action" when the CLI runs inside the oasdiff GitHub
	// Action (which sets it in its image), empty for a plain CLI invocation. It
	// lets the review page tailor an upgrade nudge (bump the action vs set up
	// the action).
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty"`
}

// Change is one manifest entry sent alongside the encrypted bundle in
// cleartext: a change's fingerprint (see checker.Fingerprint) and
// its level, so a server can track per-change state without reading the bundle.
type Change struct {
	Fingerprint string `json:"fingerprint" yaml:"fingerprint"`
	Level       int    `json:"level" yaml:"level"`
}

// Manifest builds the {fingerprint, level} manifest. Fingerprints use
// checker.Fingerprint, so they match the ones in the encrypted changelog the
// bundle carries and the fingerprints on each Block.
func Manifest(changes checker.Changes) []Change {
	manifest := make([]Change, 0, len(changes))
	for _, change := range changes {
		manifest = append(manifest, Change{
			Fingerprint: checker.Fingerprint(change),
			Level:       int(change.GetLevel()),
		})
	}
	return manifest
}
