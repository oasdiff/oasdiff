/*
Package diff calculates the difference between two OpenAPI specifications.

# Overview

The diff package compares OpenAPI 3.x specifications and produces a structured diff report
describing all changes. It supports OpenAPI 3.0 and 3.1, including JSON Schema 2020-12 keywords.

# Usage

The main entry points are Get and GetWithOperationsSourcesMap:

	config := diff.NewConfig()
	diffReport, err := diff.Get(config, spec1, spec2)

	// Or with operation source tracking for better error messages:
	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(config, specInfo1, specInfo2)

# Configuration

Config controls diff behavior:
  - MatchPath/UnmatchPath: filter paths by regex pattern
  - FilterExtension: filter by x- extension presence
  - PathPrefixBase/PathPrefixRevision: add prefix to paths before comparison
  - PathStripPrefixBase/PathStripPrefixRevision: strip prefix from paths before comparison
  - ExcludeElements: skip comparing certain elements (examples, description, title, summary, extensions, endpoints)
  - ExcludeExtensions: skip specific x- extensions
  - IncludePathParams: include path parameter names in endpoint identity

# Diff Structure

The Diff type contains nested diff objects for each OpenAPI component:
  - PathsDiff: changes to path items and operations
  - WebhooksDiff: changes to webhooks (OpenAPI 3.1)
  - ComponentsDiff: changes to reusable components (schemas, parameters, responses, etc.)
  - InfoDiff, SecurityDiff, ServersDiff, TagsDiff: other top-level changes

Each diff type follows a consistent pattern with Added, Deleted, and Modified fields.
Modified entries contain detailed nested diffs showing exactly what changed.

# Schema Diffing

SchemaDiff handles JSON Schema comparison including:
  - Type changes, format, pattern, enum values
  - Numeric constraints (min, max, multipleOf)
  - String constraints (minLength, maxLength)
  - Array constraints (minItems, maxItems, uniqueItems)
  - Object constraints (required, properties, additionalProperties)
  - Composition (allOf, oneOf, anyOf, not)
  - JSON Schema 2020-12: $defs, if/then/else, dependentSchemas, prefixItems, contains, etc.

# References

OpenAPI $ref references should be resolved before diffing. The load package resolves
refs automatically. For manually loaded specs, use openapi3.Loader.ResolveRefsIn.
*/
package diff
