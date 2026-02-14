package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxDecreasedId                      = "request-body-max-decreased"
	RequestBodyMaxIncreasedId                      = "request-body-max-increased"
	RequestPropertyMaxDecreasedId                  = "request-property-max-decreased"
	RequestReadOnlyPropertyMaxDecreasedId          = "request-read-only-property-max-decreased"
	RequestPropertyMaxIncreasedId                  = "request-property-max-increased"
	RequestBodyExclusiveMaxDecreasedId             = "request-body-exclusive-max-decreased"
	RequestBodyExclusiveMaxIncreasedId             = "request-body-exclusive-max-increased"
	RequestPropertyExclusiveMaxDecreasedId         = "request-property-exclusive-max-decreased"
	RequestReadOnlyPropertyExclusiveMaxDecreasedId = "request-read-only-property-exclusive-max-decreased"
	RequestPropertyExclusiveMaxIncreasedId         = "request-property-exclusive-max-increased"
)

func RequestPropertyMaxDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MaxDiff != nil {
					maxDiff := mediaTypeDiff.SchemaDiff.MaxDiff
					if maxDiff.From != nil &&
						maxDiff.To != nil {
						if IsDecreasedValue(maxDiff) {
							result = append(result, NewApiChange(
								RequestBodyMaxDecreasedId,
								config,
								[]any{maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestBodyMaxIncreasedId,
								config,
								[]any{maxDiff.From, maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}
					}
				}
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.ExclusiveMaxDiff != nil {
					exMaxDiff := mediaTypeDiff.SchemaDiff.ExclusiveMaxDiff
					if exMaxDiff.From != nil &&
						exMaxDiff.To != nil {
						if IsDecreasedValue(exMaxDiff) {
							result = append(result, NewApiChange(
								RequestBodyExclusiveMaxDecreasedId,
								config,
								[]any{exMaxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestBodyExclusiveMaxIncreasedId,
								config,
								[]any{exMaxDiff.From, exMaxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						maxDiff := propertyDiff.MaxDiff
						if maxDiff == nil {
							return
						}
						if maxDiff.From == nil ||
							maxDiff.To == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						id := RequestPropertyMaxDecreasedId
						if propertyDiff.Revision.ReadOnly {
							id = RequestReadOnlyPropertyMaxDecreasedId
						}

						if IsDecreasedValue(maxDiff) {
							result = append(result, NewApiChange(
								id,
								config,
								[]any{propName, maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyMaxIncreasedId,
								config,
								[]any{propName, maxDiff.From, maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}

					})

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						exMaxDiff := propertyDiff.ExclusiveMaxDiff
						if exMaxDiff == nil {
							return
						}
						if exMaxDiff.From == nil ||
							exMaxDiff.To == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						id := RequestPropertyExclusiveMaxDecreasedId
						if propertyDiff.Revision.ReadOnly {
							id = RequestReadOnlyPropertyExclusiveMaxDecreasedId
						}

						if IsDecreasedValue(exMaxDiff) {
							result = append(result, NewApiChange(
								id,
								config,
								[]any{propName, exMaxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyExclusiveMaxIncreasedId,
								config,
								[]any{propName, exMaxDiff.From, exMaxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
