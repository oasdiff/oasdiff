package formatters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
)

type Change struct {
	Id             string          `json:"id,omitempty" yaml:"id,omitempty"`
	Text           string          `json:"text,omitempty" yaml:"text,omitempty"`
	Comment        string          `json:"comment,omitempty" yaml:"comment,omitempty"`
	Level          checker.Level   `json:"level" yaml:"level"`
	Operation      string          `json:"operation,omitempty" yaml:"operation,omitempty"`
	OperationId    string          `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Path           string          `json:"path,omitempty" yaml:"path,omitempty"`
	Source         string          `json:"source,omitempty" yaml:"source,omitempty"`
	Section        string          `json:"section,omitempty" yaml:"section,omitempty"`
	IsBreaking     bool            `json:"-" yaml:"-"`
	Attributes     map[string]any  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	BaseSource     *checker.Source `json:"baseSource,omitempty" yaml:"baseSource,omitempty"`
	RevisionSource *checker.Source `json:"revisionSource,omitempty" yaml:"revisionSource,omitempty"`
	Fingerprint    string          `json:"fingerprint,omitempty" yaml:"fingerprint,omitempty"`
}

type Changes []Change

func NewChanges(originalChanges checker.Changes, l checker.Localizer) Changes {
	changes := make(Changes, len(originalChanges))
	for i, change := range originalChanges {
		id := change.GetId()
		operation := change.GetOperation()
		path := change.GetPath()
		changes[i] = Change{
			Section:        change.GetSection(),
			Id:             id,
			Text:           change.GetUncolorizedText(l),
			Comment:        change.GetComment(l),
			Level:          change.GetLevel(),
			Operation:      operation,
			OperationId:    change.GetOperationId(),
			Path:           path,
			Source:         change.GetSource(),
			Attributes:     change.GetAttributes(),
			BaseSource:     change.GetBaseSource(),
			RevisionSource: change.GetRevisionSource(),
			Fingerprint:    computeFingerprint(id, operation, path, change.GetArgs()),
		}
	}
	return changes
}

// computeFingerprint produces a short, stable identifier for a change. It is
// used to match the same logical change across spec versions (carry-forward of
// review state in oasdiff-service) and to look up review records at view time.
//
// Inputs are the structured rule arguments rather than the rendered text:
// rendered text varies with locale and copy edits to the message templates,
// which would silently invalidate every stored fingerprint. The args carry the
// same per-change disambiguation power without that fragility.
func computeFingerprint(id, operation, path string, args []any) string {
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
