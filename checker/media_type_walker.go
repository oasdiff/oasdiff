package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

// mediaTypeInfo is the per-media-type plumbing delivered by the walker
// functions in this file to their processor callbacks. It bundles what
// every body-level / response-body checker needs (path, method, operation
// diff, media type plus its formatted detail string) with the schema diff
// the processor will actually examine.
//
// responseStatus is empty for request-body walkers and carries the status
// code (e.g. "200") for response-body walkers.
//
// The unexported config and operationsSources fields back the newChange
// helper so callers do not have to re-thread them at every emission site.
type mediaTypeInfo struct {
	path             string
	method           string
	responseStatus   string
	operationItem    *diff.MethodDiff
	mediaType        string
	mediaTypeDetails string
	schemaDiff       *diff.SchemaDiff

	config            *Config
	operationsSources *diff.OperationsSourcesMap
}

// newChange constructs an ApiChange pre-filled with the plumbing carried
// by info: config, operationsSources, the revision operation, method,
// path, and the media-type detail string attached via WithDetails. The
// caller supplies the check ID, args, and comment; the returned ApiChange
// can be further decorated (typically with WithSources(baseSource,
// revisionSource), or with a further WithDetails(...) to override the
// auto-attached media-type details when a check needs a combined string).
//
// Args usually contain the change-specific payload (added/deleted names,
// property names). Response-body checkers conventionally include
// info.responseStatus in args themselves; this helper does not append it
// automatically because each check ID's args shape is fixed at the
// localization layer.
func (info mediaTypeInfo) newChange(id string, args []any, comment string) ApiChange {
	return NewApiChange(
		id,
		info.config,
		args,
		comment,
		info.operationsSources,
		info.operationItem.Revision,
		info.method,
		info.path,
	).WithDetails(info.mediaTypeDetails)
}

// walkProperties walks every modified property under info.schemaDiff and
// invokes processor with a propertyInfo for each. Delegates to
// CheckModifiedPropertiesDiff, so the recursion covers AllOf / AnyOf /
// OneOf / Items / PatternProperties / DependentSchemas and the OpenAPI 3.1
// sub-schema fields exactly as that primitive does.
//
// p.newChange in the processor produces an ApiChange with the same
// plumbing as info.newChange (config, operationsSources, operation,
// method, path, mediaTypeDetails); the caller chains
// WithSources(propBaseSource, propRevisionSource) with the property-
// specific sources.
func (info mediaTypeInfo) walkProperties(processor func(p propertyInfo)) {
	CheckModifiedPropertiesDiff(info.schemaDiff, func(propertyPath, propertyName string, propertyDiff, parent *diff.SchemaDiff) {
		// The traversal also descends into single-valued sub-schemas (items,
		// not, if/then/else, contentSchema, ...). When such a sub-schema exists
		// on only one side (e.g. `items` removed in the revision), its diff has
		// a nil Base or Revision schema. The property-level checks all read
		// Base/Revision (ReadOnly, WriteOnly, Extensions, ...) and have nothing
		// actionable to say about a side that doesn't exist, so guard once here
		// rather than in every check.
		if propertyDiff == nil || propertyDiff.Base == nil || propertyDiff.Revision == nil {
			return
		}
		processor(propertyInfo{
			mediaTypeInfo: info,
			propertyPath:  propertyPath,
			propertyName:  propertyName,
			propertyDiff:  propertyDiff,
			parent:        parent,
		})
	})
}

// propertyInfo is the per-property plumbing delivered by
// mediaTypeInfo.walkProperties. It embeds the originating mediaTypeInfo
// so p.newChange resolves to the body-level helper via field promotion:
// the same plumbing is pre-filled, and the same mediaTypeDetails is
// auto-attached. Override Details with a further WithDetails(...) when a
// property check needs a combined detail string (e.g. deprecation +
// media-type).
//
// propertyPath, propertyName, and propertyDiff are the same triple
// CheckModifiedPropertiesDiff passes to its processor; parent is the
// containing schema diff.
type propertyInfo struct {
	mediaTypeInfo
	propertyPath string
	propertyName string
	propertyDiff *diff.SchemaDiff
	parent       *diff.SchemaDiff
}

// modifiedSchemaPresentBothSides reports whether a media type's schema changed
// with a schema present on BOTH sides. A schema added (Base nil) or removed
// (Revision nil) on an existing media type is not a modification of an existing
// schema, and the per-schema-change checks built on these walkers dereference
// Base/Revision unconditionally, so walking those cases nil-panics (#1047).
func modifiedSchemaPresentBothSides(d *diff.MediaTypeDiff) bool {
	return d.SchemaDiff != nil && d.SchemaDiff.Base != nil && d.SchemaDiff.Revision != nil
}

// walkModifiedRequestBodySchemas invokes processor once for each modified
// request-body media type across the diff, with the per-media-type schema
// diff and the plumbing needed to emit ApiChange values. Skips media types
// whose schema is absent on either side (added or removed; see
// modifiedSchemaPresentBothSides).
func walkModifiedRequestBodySchemas(
	diffReport *diff.Diff,
	operationsSources *diff.OperationsSourcesMap,
	config *Config,
	processor func(info mediaTypeInfo),
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
				if !modifiedSchemaPresentBothSides(mediaTypeDiff) {
					continue
				}
				processor(mediaTypeInfo{
					path:              path,
					method:            method,
					operationItem:     operationItem,
					mediaType:         mediaType,
					mediaTypeDetails:  formatMediaTypeDetails(mediaType, len(modifiedMediaTypes)),
					schemaDiff:        mediaTypeDiff.SchemaDiff,
					config:            config,
					operationsSources: operationsSources,
				})
			}
		}
	}
}

// walkModifiedResponseSchemas mirrors walkModifiedRequestBodySchemas for
// response bodies. The processor receives an info with responseStatus set
// to the response status code (e.g. "200"). Skips media types whose
// schema is absent on either side (see modifiedSchemaPresentBothSides).
func walkModifiedResponseSchemas(
	diffReport *diff.Diff,
	operationsSources *diff.OperationsSourcesMap,
	config *Config,
	processor func(info mediaTypeInfo),
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
					if !modifiedSchemaPresentBothSides(mediaTypeDiff) {
						continue
					}
					processor(mediaTypeInfo{
						path:              path,
						method:            method,
						responseStatus:    responseStatus,
						operationItem:     operationItem,
						mediaType:         mediaType,
						mediaTypeDetails:  formatMediaTypeDetails(mediaType, len(modifiedMediaTypes)),
						schemaDiff:        mediaTypeDiff.SchemaDiff,
						config:            config,
						operationsSources: operationsSources,
					})
				}
			}
		}
	}
}
