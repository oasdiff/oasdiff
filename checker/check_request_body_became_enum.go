package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecameEnumId = "request-body-became-enum"
)

func RequestBodyBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified

			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}
				if schemaDiff := mediaTypeDiff.SchemaDiff; schemaDiff.EnumDiff == nil || !schemaDiff.EnumDiff.EnumAdded {
					continue
				}
				result = append(result, NewApiChange(
					RequestBodyBecameEnumId,
					config,
					nil,
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
			}
		}
	}
	return result
}
