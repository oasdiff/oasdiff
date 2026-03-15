/*
Package formatters transforms oasdiff output into various formats for display and integration.

# Overview

The formatters package provides a unified interface for rendering diff results, changelog,
and breaking changes in multiple output formats suitable for different use cases.

# Usage

Get a formatter and render output:

	formatter, err := formatters.Lookup("json", formatters.FormatterOpts{Language: "en"})
	output, err := formatter.RenderChangelog(changes, formatters.RenderOpts{}, baseVersion, revisionVersion)

# Available Formats

  - yaml: YAML output for programmatic consumption
  - json: JSON output for programmatic consumption
  - text: Human-readable text output for terminals
  - markup/markdown: Markdown format for documentation
  - singleline: One-line-per-change format for parsing
  - html: HTML format for web display
  - githubactions: GitHub Actions workflow command format (::error, ::warning)
  - junit: JUnit XML format for CI/CD test reporting

# Formatter Interface

All formatters implement the Formatter interface with methods:
  - RenderDiff: render a diff report
  - RenderSummary: render a summary of changes
  - RenderChangelog: render breaking changes and changelog
  - RenderChecks: render available check rules
  - RenderFlatten: render a flattened spec

# Localization

Formatters support localization through the Language option in FormatterOpts.
Change messages are translated using the checker/localizations package.

# Output Types

Use SupportedOutputs() to check which output types a formatter supports:
  - OutputDiff, OutputSummary, OutputChangelog, OutputChecks, OutputFlatten
*/
package formatters
