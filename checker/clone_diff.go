package checker

import (
	"maps"
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

// cloneDiffForCheck returns a copy of diffReport in which every part the
// check pipeline writes to is a separate allocation from the caller's.
// Everything else (WebhooksDiff, ComponentsDiff, the MethodDiff values, and
// the contexts, which are only read or replaced) is shared by reference.
func cloneDiffForCheck(d *diff.Diff) *diff.Diff {
	cloned := *d
	cloned.PathsDiff = clonePathsDiff(d.PathsDiff)
	return &cloned
}

// clonePathsDiff copies the Deleted slice, which the stability filter
// truncates in place, and the Modified map, which receives the webhook
// entries from mergeWebhookOperationsIntoPathsDiff.
func clonePathsDiff(pathsDiff *diff.PathsDiff) *diff.PathsDiff {
	if pathsDiff == nil {
		return nil
	}
	cloned := *pathsDiff
	cloned.Deleted = slices.Clone(pathsDiff.Deleted)
	if pathsDiff.Modified != nil {
		cloned.Modified = make(diff.ModifiedPaths, len(pathsDiff.Modified))
		for path, pathDiff := range pathsDiff.Modified {
			cloned.Modified[path] = clonePathDiff(pathDiff)
		}
	}
	return &cloned
}

func clonePathDiff(pathDiff *diff.PathDiff) *diff.PathDiff {
	cloned := *pathDiff
	cloned.OperationsDiff = cloneOperationsDiff(pathDiff.OperationsDiff)
	return &cloned
}

// cloneOperationsDiff copies the Deleted slice, which the stability filter
// truncates in place, and the Modified map, which it deletes keys from.
func cloneOperationsDiff(operationsDiff *diff.OperationsDiff) *diff.OperationsDiff {
	if operationsDiff == nil {
		return nil
	}
	cloned := *operationsDiff
	cloned.Deleted = slices.Clone(operationsDiff.Deleted)
	if operationsDiff.Modified != nil {
		cloned.Modified = make(diff.ModifiedOperations, len(operationsDiff.Modified))
		maps.Copy(cloned.Modified, operationsDiff.Modified)
	}
	return &cloned
}
