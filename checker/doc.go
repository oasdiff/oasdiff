/*
Package checker detects breaking changes and generates changelog messages between OpenAPI specifications.

# Overview

The checker analyzes a diff report (from the diff package) and applies a set of rules to detect
breaking changes and other notable differences. It supports both path operations and webhooks.

# Usage

The main entry point is CheckBackwardCompatibility:

	diffReport, _ := diff.Get(config, spec1, spec2)
	changes := checker.CheckBackwardCompatibility(checker.NewConfig(checker.GetAllRules()), diffReport, operationsSources)

# Configuration

Config controls which checks run and their severity levels:
  - Checks: the list of BackwardCompatibilityCheck functions to run
  - LogLevels: override default severity (ERR, WARN, INFO) for specific rule IDs
  - MinSunsetBetaDays/MinSunsetStableDays: minimum deprecation periods

Use NewConfig with GetAllRules() for all checks, or GetDefaultRules() for breaking changes only.

# Severity Levels

Each rule has a default severity level:
  - ERR: breaking change that will break existing clients
  - WARN: potentially breaking change that may affect some clients
  - INFO: non-breaking change for changelog purposes

# Rules

Rules are defined in rules.go with metadata including:
  - Direction: whether the rule applies to requests, responses, or neither
  - Location: body, parameters, properties, headers, security, or components
  - Action: add, remove, change, generalize, specialize, increase, decrease, set

This metadata enables filtering and categorization of changes.

# Webhooks

Webhooks are handled by merging modified webhooks into PathsDiff with a "webhook:" prefix,
allowing all path/operation rules to apply without duplication. Added/deleted webhooks are
handled separately by WebhookUpdatedCheck since PathsDiff.Added/Deleted require lookup in
openapi3.Paths objects that don't contain webhooks.

# Localization

Change messages support localization via the localizations package (generated from
localizations_src). Messages use format strings with placeholders for dynamic values.
*/
package checker
