package formatters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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
		text := change.GetUncolorizedText(l)
		changes[i] = Change{
			Section:        change.GetSection(),
			Id:             id,
			Text:           text,
			Comment:        change.GetComment(l),
			Level:          change.GetLevel(),
			Operation:      operation,
			OperationId:    change.GetOperationId(),
			Path:           path,
			Source:         change.GetSource(),
			Attributes:     change.GetAttributes(),
			BaseSource:     change.GetBaseSource(),
			RevisionSource: change.GetRevisionSource(),
			Fingerprint:    computeFingerprint(id, operation, path, text),
		}
	}
	return changes
}

func computeFingerprint(id, operation, path, text string) string {
	h := fmt.Sprintf("%s:%s:%s:%s", id, operation, path, text)
	sum := sha256.Sum256([]byte(h))
	return hex.EncodeToString(sum[:])[:12]
}
