package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDiscriminatorAddedId                   = "response-body-discriminator-added"
	ResponseBodyDiscriminatorRemovedId                 = "response-body-discriminator-removed"
	ResponseBodyDiscriminatorPropertyNameChangedId     = "response-body-discriminator-property-name-changed"
	ResponseBodyDiscriminatorMappingAddedId            = "response-body-discriminator-mapping-added"
	ResponseBodyDiscriminatorMappingDeletedId          = "response-body-discriminator-mapping-deleted"
	ResponseBodyDiscriminatorMappingChangedId          = "response-body-discriminator-mapping-changed"
	ResponsePropertyDiscriminatorAddedId               = "response-property-discriminator-added"
	ResponsePropertyDiscriminatorRemovedId             = "response-property-discriminator-removed"
	ResponsePropertyDiscriminatorPropertyNameChangedId = "response-property-discriminator-property-name-changed"
	ResponsePropertyDiscriminatorMappingAddedId        = "response-property-discriminator-mapping-added"
	ResponsePropertyDiscriminatorMappingDeletedId      = "response-property-discriminator-mapping-deleted"
	ResponsePropertyDiscriminatorMappingChangedId      = "response-property-discriminator-mapping-changed"
)

func ResponseDiscriminatorUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}

		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}

			appendResultItem := func(messageId string, a ...any) {
				baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
				result = append(result, NewApiChange(
					messageId,
					config,
					a,
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}

			for responseStatus, responsesDiff := range operationItem.ResponsesDiff.Modified {
				if responsesDiff.ContentDiff == nil || responsesDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responsesDiff.ContentDiff.MediaTypeModified
				for _, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					processDiscriminatorDiff(
						mediaTypeDiff.SchemaDiff.DiscriminatorDiff,
						responseStatus,
						"",
						appendResultItem)

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							processDiscriminatorDiff(
								propertyDiff.DiscriminatorDiff,
								responseStatus,
								propertyFullName(propertyPath, propertyName),
								appendResultItem)
						})
				}
			}
		}
	}
	return result
}

func processDiscriminatorDiff(
	discriminatorDiff *diff.DiscriminatorDiff,
	responseStatus string,
	propertyName string,
	appendResultItem func(messageId string, a ...any)) {

	if discriminatorDiff == nil {
		return
	}

	messageIdPrefix := "response-body-discriminator"
	if propertyName != "" {
		messageIdPrefix = "response-property-discriminator"
	}

	if discriminatorDiff.Added {
		if propertyName == "" {
			appendResultItem(messageIdPrefix+"-added", responseStatus)
		} else {
			appendResultItem(messageIdPrefix+"-added", propertyName, responseStatus)
		}
		return
	}
	if discriminatorDiff.Deleted {
		if propertyName == "" {
			appendResultItem(messageIdPrefix+"-removed", responseStatus)
		} else {
			appendResultItem(messageIdPrefix+"-removed", propertyName, responseStatus)
		}
		return
	}

	if discriminatorDiff.PropertyNameDiff != nil {
		if propertyName == "" {
			appendResultItem(messageIdPrefix+"-property-name-changed",
				discriminatorDiff.PropertyNameDiff.From,
				discriminatorDiff.PropertyNameDiff.To,
				responseStatus)
		} else {
			appendResultItem(messageIdPrefix+"-property-name-changed",
				propertyName,
				discriminatorDiff.PropertyNameDiff.From,
				discriminatorDiff.PropertyNameDiff.To,
				responseStatus)
		}
	}

	if discriminatorDiff.MappingDiff != nil {
		if len(discriminatorDiff.MappingDiff.Added) > 0 {
			if propertyName == "" {
				appendResultItem(messageIdPrefix+"-mapping-added",
					discriminatorDiff.MappingDiff.Added,
					responseStatus)
			} else {
				appendResultItem(messageIdPrefix+"-mapping-added",
					discriminatorDiff.MappingDiff.Added,
					propertyName,
					responseStatus)
			}
		}

		if len(discriminatorDiff.MappingDiff.Deleted) > 0 {
			if propertyName == "" {
				appendResultItem(messageIdPrefix+"-mapping-deleted",
					discriminatorDiff.MappingDiff.Deleted,
					responseStatus)
			} else {
				appendResultItem(messageIdPrefix+"-mapping-deleted",
					discriminatorDiff.MappingDiff.Deleted,
					propertyName,
					responseStatus)
			}
		}

		for k, v := range discriminatorDiff.MappingDiff.Modified {
			if propertyName == "" {
				appendResultItem(messageIdPrefix+"-mapping-changed",
					k,
					v.From,
					v.To,
					responseStatus)
			} else {
				appendResultItem(messageIdPrefix+"-mapping-changed",
					k,
					v.From,
					v.To,
					propertyName,
					responseStatus)

			}
		}
	}
}
