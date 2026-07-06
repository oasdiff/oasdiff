package formatters

import (
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
			Attributes:     change.GetAttributes(),
			BaseSource:     change.GetBaseSource(),
			RevisionSource: change.GetRevisionSource(),
			Fingerprint:    checker.ComputeFingerprint(id, operation, path, change.GetArgs()),
		}
	}
	return changes
}
