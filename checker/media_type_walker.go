package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

// MediaTypeChangeCtx is the per-media-type context delivered by the walker
// functions in this file to their processor callbacks. It bundles the
// plumbing every body-level / response-body checker needs (path, method,
// operation diff, media type plus its formatted detail string) with the
// schema diff the processor will actually examine.
//
// ResponseStatus is empty for request-body walkers and carries the status
// code (e.g. "200") for response-body walkers.
//
// The unexported fields carry the *Config and the OperationsSourcesMap
// through so the NewChange helper can construct ApiChange values without
// the caller having to re-thread them at every emission site.
type MediaTypeChangeCtx struct {
	Path             string
	Method           string
	ResponseStatus   string
	OperationItem    *diff.MethodDiff
	MediaType        string
	MediaTypeDetails string
	SchemaDiff       *diff.SchemaDiff

	config            *Config
	operationsSources *diff.OperationsSourcesMap
}

// NewChange constructs an ApiChange pre-filled with the plumbing carried by
// ctx: config, operationsSources, the revision operation, method, path, and
// the media-type detail string attached via WithDetails. The caller supplies
// the check ID, args, and comment; the returned ApiChange can be further
// decorated (typically with WithSources(baseSource, revisionSource)).
//
// Args usually contain the change-specific payload (added/deleted names,
// property names). Response-body checkers conventionally include
// ctx.ResponseStatus in args themselves; this helper does not append it
// automatically because each check ID's args shape is fixed at the
// localization layer.
func (ctx MediaTypeChangeCtx) NewChange(id string, args []any, comment string) ApiChange {
	return NewApiChange(
		id,
		ctx.config,
		args,
		comment,
		ctx.operationsSources,
		ctx.OperationItem.Revision,
		ctx.Method,
		ctx.Path,
	).WithDetails(ctx.MediaTypeDetails)
}

// WalkModifiedRequestBodySchemas invokes processor once for each modified
// request-body media type across the diff, with the per-media-type schema
// diff and the plumbing needed to emit ApiChange values. Skips media types
// whose SchemaDiff is nil.
//
// This is the body-level walker. Callers that also need to iterate the
// schema's properties keep using CheckModifiedPropertiesDiff inside the
// processor; this walker intentionally does not recurse into properties.
func WalkModifiedRequestBodySchemas(
	diffReport *diff.Diff,
	operationsSources *diff.OperationsSourcesMap,
	config *Config,
	processor func(ctx MediaTypeChangeCtx),
) {
	if diffReport == nil || diffReport.PathsDiff == nil {
		return
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for method, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}
			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}
				processor(MediaTypeChangeCtx{
					Path:              path,
					Method:            method,
					OperationItem:     operationItem,
					MediaType:         mediaType,
					MediaTypeDetails:  formatMediaTypeDetails(mediaType, len(modifiedMediaTypes)),
					SchemaDiff:        mediaTypeDiff.SchemaDiff,
					config:            config,
					operationsSources: operationsSources,
				})
			}
		}
	}
}

// WalkModifiedResponseSchemas mirrors WalkModifiedRequestBodySchemas for
// response bodies. The processor receives a ctx with ResponseStatus set to
// the response status code (e.g. "200"). Skips media types whose SchemaDiff
// is nil.
func WalkModifiedResponseSchemas(
	diffReport *diff.Diff,
	operationsSources *diff.OperationsSourcesMap,
	config *Config,
	processor func(ctx MediaTypeChangeCtx),
) {
	if diffReport == nil || diffReport.PathsDiff == nil {
		return
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for method, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil || responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}
					processor(MediaTypeChangeCtx{
						Path:              path,
						Method:            method,
						ResponseStatus:    responseStatus,
						OperationItem:     operationItem,
						MediaType:         mediaType,
						MediaTypeDetails:  formatMediaTypeDetails(mediaType, len(modifiedMediaTypes)),
						SchemaDiff:        mediaTypeDiff.SchemaDiff,
						config:            config,
						operationsSources: operationsSources,
					})
				}
			}
		}
	}
}
