package checker

import (
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/internal/oasversion"
)

// dialectsOf returns the dialects of the two documents the diff compares.
func dialectsOf(diffReport *diff.Diff) oasversion.Dialects {
	return oasversion.Dialects{
		Base:     oasversion.Of(diffReport.BaseOpenAPI),
		Revision: oasversion.Of(diffReport.RevisionOpenAPI),
	}
}
