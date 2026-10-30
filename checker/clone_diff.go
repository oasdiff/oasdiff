package checker

import (
	"maps"
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

// cloneDiffForCheck returns a copy of diffReport in which every part the
// check pipeline writes to is a separate allocation from the caller's.
// Everything else (WebhooksDiff, ComponentsDiff, and the MethodDiff values,
// which are only read) is shared by reference.
func cloneDiffForCheck(d *diff.Diff) *diff.Diff {
	cloned := *d
	cloned.SecurityContext = cloneSecurityContext(d.SecurityContext)
	cloned.ParametersContext = cloneParametersContext(d.ParametersContext)
	cloned.PathsDiff = clonePathsDiff(d.PathsDiff)
	return &cloned
}

// cloneSecurityContext copies the Operations slice, which
// filterSecurityContextBelowStabilityThreshold deletes pairs from in place.
func cloneSecurityContext(securityContext *diff.SecurityContext) *diff.SecurityContext {
	if securityContext == nil {
		return nil
	}
	cloned := *securityContext
	cloned.Operations = slices.Clone(securityContext.Operations)
	return &cloned
}

// cloneParametersContext copies the slices, which
// filterParametersContextBelowStabilityThreshold deletes operations from in
// place.
func cloneParametersContext(parametersContext *diff.ParametersContext) *diff.ParametersContext {
	if parametersContext == nil {
		return nil
	}
	cloned := *parametersContext
	cloned.Operations = slices.Clone(parametersContext.Operations)
	cloned.Webhooks = slices.Clone(parametersContext.Webhooks)
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
