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

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "discriminator")
		appendResultItem := func(messageId string, a ...any) {
			result = append(result, info.newChange(messageId, a, "").
				WithSources(baseSource, revisionSource))
		}

		processDiscriminatorDiff(
			info.schemaDiff.DiscriminatorDiff,
			info.responseStatus,
			"",
			appendResultItem)

		info.walkProperties(func(p propertyInfo) {
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "discriminator")
			propAppendResultItem := func(messageId string, a ...any) {
				result = append(result, p.newChange(messageId, a, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			processDiscriminatorDiff(
				p.propertyDiff.DiscriminatorDiff,
				info.responseStatus,
				propertyFullName(p.propertyPath, p.propertyName),
				propAppendResultItem)
		})
	})

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
