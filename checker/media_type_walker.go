package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

// mediaTypeInfo is what the walkers in this file hand their processors.
//
// responseStatus is empty for request bodies and carries the status code for
// responses.
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

// newChange returns an ApiChange with info's plumbing pre-filled, for the
// caller to decorate further.
//
// It does not append responseStatus to args, though response checks generally
// pass it: each check ID's args shape is fixed at the localization layer, so
// the check owns that decision.
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
	).WithSchema(info.schemaDiff).WithDetails(info.mediaTypeDetails).
		WithDisclaimers(allOfDisclaimers(false, info.schemaDiff))
}

// allOfDisclaimers reports the conditions that hold for a change at this
// location. An allOf surviving into the diff is itself the evidence: had the
// branches been flattened, there would be none left to compare.
func allOfDisclaimers(underAllOf bool, schemaDiff *diff.SchemaDiff) []Disclaimer {
	if schemaDiff != nil && schemaDiff.AllOfDiff != nil {
		underAllOf = true
	}
	if !underAllOf {
		return nil
	}
	return []Disclaimer{DisclaimerAllOfNotFlattened}
}

// walkProperties invokes processor for every modified property under
// info.schemaDiff. The recursion is checkModifiedPropertiesDiff's, so sub-schema
// coverage stays whatever that primitive does.
func (info mediaTypeInfo) walkProperties(processor func(p propertyInfo)) {
	if info.schemaDiff == nil {
		return
	}
	subschemaWalk{enter: func(propertyPath, propertyName string, propertyDiff, parent *diff.SchemaDiff, underAllOf bool) {
		// A single-valued sub-schema present on one side only (items removed,
		// say) has a nil Base or Revision. Every property check reads both and
		// has nothing to say about a side that does not exist, so guard here
		// rather than in each of them.
		if propertyDiff == nil || propertyDiff.Base == nil || propertyDiff.Revision == nil {
			return
		}
		processor(propertyInfo{
			mediaTypeInfo: info,
			underAllOf:    underAllOf,
			propertyPath:  propertyPath,
			propertyName:  propertyName,
			propertyDiff:  propertyDiff,
			parent:        parent,
		})
	}}.walk("", "", info.schemaDiff, nil, false)
}

// propertyInfo is what walkProperties hands its processor. It embeds
// mediaTypeInfo so the body-level helpers are promoted onto it.
type propertyInfo struct {
	mediaTypeInfo
	propertyPath string
	propertyName string
	underAllOf   bool
	propertyDiff *diff.SchemaDiff
	parent       *diff.SchemaDiff
}

// newChange shadows the promoted body-level helper so the claim decision is
// made against the property's own schema diff (WithSchema recomputes claimed,
// so the second call overrides the body-level decision).
func (p propertyInfo) newChange(id string, args []any, comment string) ApiChange {
	return p.mediaTypeInfo.newChange(id, args, comment).WithSchema(p.propertyDiff).
		WithDisclaimers(allOfDisclaimers(p.underAllOf, p.propertyDiff)).
		WithGuards(propertyGuards(p.propertyDiff))
}

// modifiedSchemaPresentBothSides reports whether a schema changed on both
// sides. The checks built on these walkers dereference Base and Revision
// unconditionally, so a one-sided schema would panic; those cases are
// MediaTypeSchemaExistenceCheck's to report.
func modifiedSchemaPresentBothSides(d *diff.SchemaDiff) bool {
	return d != nil && d.Base != nil && d.Revision != nil
}

// itemSchemaDetail marks a change under a media type's itemSchema (OpenAPI 3.2,
// one item of a streamed body). Unconditional, unlike the media-type detail:
// these are the body-schema checks, whose messages talk about the body, so
// without it a streamed item reads as the body itself.
const itemSchemaDetail = "(item schema)"

// walkMediaTypeSchemas invokes processor for each of a media type's changed
// schemas: the body schema, then the OpenAPI 3.2 itemSchema. Checks built on
// the walkers cover streamed items without knowing about them, which is what
// keeps the two in step as checks are added.
func walkMediaTypeSchemas(mediaTypeDiff *diff.MediaTypeDiff, info mediaTypeInfo, processor func(info mediaTypeInfo)) {
	mediaTypeDetails := info.mediaTypeDetails

	if modifiedSchemaPresentBothSides(mediaTypeDiff.SchemaDiff) {
		info.schemaDiff = mediaTypeDiff.SchemaDiff
		processor(info)
	}

	if modifiedSchemaPresentBothSides(mediaTypeDiff.ItemSchemaDiff) {
		info.schemaDiff = mediaTypeDiff.ItemSchemaDiff
		info.mediaTypeDetails = combineDetails(mediaTypeDetails, itemSchemaDetail)
		processor(info)
	}
}

// walkModifiedRequestBodySchemas invokes processor for each modified
// request-body media type.
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
				walkMediaTypeSchemas(mediaTypeDiff, mediaTypeInfo{
					path:              path,
					method:            method,
					operationItem:     operationItem,
					mediaType:         mediaType,
					mediaTypeDetails:  formatMediaTypeDetails(mediaType, len(modifiedMediaTypes)),
					config:            config,
					operationsSources: operationsSources,
				}, processor)
			}
		}
	}
}

// walkModifiedResponseSchemas mirrors walkModifiedRequestBodySchemas for
// response bodies, setting responseStatus on the info it passes.
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
					walkMediaTypeSchemas(mediaTypeDiff, mediaTypeInfo{
						path:              path,
						method:            method,
						responseStatus:    responseStatus,
						operationItem:     operationItem,
						mediaType:         mediaType,
						mediaTypeDetails:  formatMediaTypeDetails(mediaType, len(modifiedMediaTypes)),
						config:            config,
						operationsSources: operationsSources,
					}, processor)
				}
			}
		}
	}
}
