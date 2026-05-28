package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDiscriminatorAddedId                   = "request-body-discriminator-added"
	RequestBodyDiscriminatorRemovedId                 = "request-body-discriminator-removed"
	RequestBodyDiscriminatorPropertyNameChangedId     = "request-body-discriminator-property-name-changed"
	RequestBodyDiscriminatorMappingAddedId            = "request-body-discriminator-mapping-added"
	RequestBodyDiscriminatorMappingDeletedId          = "request-body-discriminator-mapping-deleted"
	RequestBodyDiscriminatorMappingChangedId          = "request-body-discriminator-mapping-changed"
	RequestPropertyDiscriminatorAddedId               = "request-property-discriminator-added"
	RequestPropertyDiscriminatorRemovedId             = "request-property-discriminator-removed"
	RequestPropertyDiscriminatorPropertyNameChangedId = "request-property-discriminator-property-name-changed"
	RequestPropertyDiscriminatorMappingAddedId        = "request-property-discriminator-mapping-added"
	RequestPropertyDiscriminatorMappingDeletedId      = "request-property-discriminator-mapping-deleted"
	RequestPropertyDiscriminatorMappingChangedId      = "request-property-discriminator-mapping-changed"
)

func RequestDiscriminatorUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		bodyBaseSource, bodyRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "discriminator")
		appendBodyResultItem := func(messageId string, a ...any) {
			result = append(result, info.newChange(messageId, a, "").
				WithSources(bodyBaseSource, bodyRevisionSource))
		}

		processDiscriminatorDiffForRequest(
			info.schemaDiff.DiscriminatorDiff,
			"",
			appendBodyResultItem)

		info.walkProperties(func(p propertyInfo) {
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "discriminator")
			appendPropResultItem := func(messageId string, a ...any) {
				result = append(result, p.newChange(messageId, a, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			processDiscriminatorDiffForRequest(
				p.propertyDiff.DiscriminatorDiff,
				propertyFullName(p.propertyPath, p.propertyName),
				appendPropResultItem)
		})
	})

	return result
}

func processDiscriminatorDiffForRequest(
	discriminatorDiff *diff.DiscriminatorDiff,
	propertyName string,
	appendResultItem func(messageId string, a ...any)) {

	if discriminatorDiff == nil {
		return
	}

	messageIdPrefix := "request-body-discriminator"
	if propertyName != "" {
		messageIdPrefix = "request-property-discriminator"
	}

	if discriminatorDiff.Added {
		if propertyName == "" {
			appendResultItem(messageIdPrefix + "-added")
		} else {
			appendResultItem(messageIdPrefix+"-added", propertyName)
		}
		return
	}
	if discriminatorDiff.Deleted {
		if propertyName == "" {
			appendResultItem(messageIdPrefix + "-removed")
		} else {
			appendResultItem(messageIdPrefix+"-removed", propertyName)
		}
		return
	}

	if discriminatorDiff.PropertyNameDiff != nil {
		if propertyName == "" {
			appendResultItem(messageIdPrefix+"-property-name-changed",
				discriminatorDiff.PropertyNameDiff.From,
				discriminatorDiff.PropertyNameDiff.To)
		} else {
			appendResultItem(messageIdPrefix+"-property-name-changed",
				propertyName,
				discriminatorDiff.PropertyNameDiff.From,
				discriminatorDiff.PropertyNameDiff.To)
		}
	}

	if discriminatorDiff.MappingDiff != nil {
		if len(discriminatorDiff.MappingDiff.Added) > 0 {
			if propertyName == "" {
				appendResultItem(messageIdPrefix+"-mapping-added",
					discriminatorDiff.MappingDiff.Added)
			} else {
				appendResultItem(messageIdPrefix+"-mapping-added",
					discriminatorDiff.MappingDiff.Added,
					propertyName)
			}
		}

		if len(discriminatorDiff.MappingDiff.Deleted) > 0 {
			if propertyName == "" {
				appendResultItem(messageIdPrefix+"-mapping-deleted",
					discriminatorDiff.MappingDiff.Deleted)
			} else {
				appendResultItem(messageIdPrefix+"-mapping-deleted",
					discriminatorDiff.MappingDiff.Deleted,
					propertyName)
			}
		}

		for k, v := range discriminatorDiff.MappingDiff.Modified {
			if propertyName == "" {
				appendResultItem(messageIdPrefix+"-mapping-changed",
					k,
					v.From,
					v.To)
			} else {
				appendResultItem(messageIdPrefix+"-mapping-changed",
					k,
					v.From,
					v.To,
					propertyName)

			}
		}
	}
}
