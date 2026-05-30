package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyAllOfAddedId                     = "request-body-all-of-added"
	RequestBodyAllOfRemovedId                   = "request-body-all-of-removed"
	RequestPropertyAllOfAddedId                 = "request-property-all-of-added"
	RequestPropertyAllOfRemovedId               = "request-property-all-of-removed"
	RequestBodyAllOfAddedAnnotationOnlyId       = "request-body-all-of-added-annotation-only"
	RequestBodyAllOfRemovedAnnotationOnlyId     = "request-body-all-of-removed-annotation-only"
	RequestPropertyAllOfAddedAnnotationOnlyId   = "request-property-all-of-added-annotation-only"
	RequestPropertyAllOfRemovedAnnotationOnlyId = "request-property-all-of-removed-annotation-only"
)

func RequestPropertyAllOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.AllOfDiff != nil {
			added, annotationOnlyAdded := splitSubschemasByAnnotationOnly(info.schemaDiff.AllOfDiff.Added, info.schemaDiff.Revision.AllOf)
			if len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, added[0].Index)
				result = append(result, info.newChange(RequestBodyAllOfAddedId, []any{added.String()}, "").
					WithSources(baseSource, revisionSource))
			}
			if len(annotationOnlyAdded) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, annotationOnlyAdded[0].Index)
				result = append(result, info.newChange(RequestBodyAllOfAddedAnnotationOnlyId, []any{annotationOnlyAdded.String()}, "").
					WithSources(baseSource, revisionSource))
			}

			deleted, annotationOnlyDeleted := splitSubschemasByAnnotationOnly(info.schemaDiff.AllOfDiff.Deleted, info.schemaDiff.Base.AllOf)
			if len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", deleted[0].Index, -1)
				result = append(result, info.newChange(RequestBodyAllOfRemovedId, []any{deleted.String()}, "").
					WithSources(baseSource, revisionSource))
			}
			if len(annotationOnlyDeleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", annotationOnlyDeleted[0].Index, -1)
				result = append(result, info.newChange(RequestBodyAllOfRemovedAnnotationOnlyId, []any{annotationOnlyDeleted.String()}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AllOfDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			added, annotationOnlyAdded := splitSubschemasByAnnotationOnly(p.propertyDiff.AllOfDiff.Added, p.propertyDiff.Revision.AllOf)
			if len(added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, added[0].Index)
				result = append(result, p.newChange(RequestPropertyAllOfAddedId, []any{added.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if len(annotationOnlyAdded) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, annotationOnlyAdded[0].Index)
				result = append(result, p.newChange(RequestPropertyAllOfAddedAnnotationOnlyId, []any{annotationOnlyAdded.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}

			deleted, annotationOnlyDeleted := splitSubschemasByAnnotationOnly(p.propertyDiff.AllOfDiff.Deleted, p.propertyDiff.Base.AllOf)
			if len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", deleted[0].Index, -1)
				result = append(result, p.newChange(RequestPropertyAllOfRemovedId, []any{deleted.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if len(annotationOnlyDeleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", annotationOnlyDeleted[0].Index, -1)
				result = append(result, p.newChange(RequestPropertyAllOfRemovedAnnotationOnlyId, []any{annotationOnlyDeleted.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
