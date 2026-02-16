/*
Package report generates basic human-readable diff reports in text and HTML formats.

# Overview

The report package provides simple, human-readable output for OpenAPI spec differences.
It focuses on common change types and is designed for quick visual review.

Note: This is a legacy package that predates the richer changelog functionality.
For full changelog with breaking change detection, localization, and multiple output
formats, use the checker package with formatters instead:

	changes := checker.CheckBackwardCompatibility(config, diffReport, operationsSources)
	formatter, _ := formatters.Lookup("text", formatters.FormatterOpts{})
	output, _ := formatter.RenderChangelog(changes, formatters.RenderOpts{}, baseVersion, revisionVersion)

# Usage

Generate a text report:

	report.GetTextReport(writer, diffReport)

Generate an HTML report:

	report.GetHTMLReport(writer, diffReport, config)

The HTMLConfig allows customization of the HTML output title.

# Output Structure

Reports display changes in a hierarchical, indented format:
  - Top-level sections for paths, components, security, etc.
  - Added/deleted/modified subsections
  - Nested details for operations, parameters, schemas

# Limitations

This package provides basic diff visualization only. It does not include:
  - Breaking change detection and severity levels
  - Localized messages
  - Rule-based change categorization

For these features, use the checker and formatters packages instead.
*/
package report
