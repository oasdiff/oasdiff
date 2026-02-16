/*
Package generator generates consistent breaking change and changelog message definitions.

# Overview

The generator package creates message IDs and localized message templates from a
declarative YAML configuration. This ensures consistency across all checker rules
and simplifies adding new change types.

# Usage

Generate messages from a YAML tree definition:

	messages, err := generator.Generate(generator.GetTree("changes.yaml"))

# YAML Structure

The input YAML defines a tree of changes with:
  - changes: the main change hierarchy (paths, operations, parameters, etc.)
  - components: reusable change definitions referenced via $ref

Each change node can have:
  - actions: map of action types (add, remove, change) to objects affected
  - nextLevel: nested changes for child elements
  - excludeFromHierarchy: whether to skip this level in message hierarchy

# Generated Output

For each action/object combination, the generator produces:
  - A kebab-case ID (e.g., "request-property-pattern-added")
  - A human-readable message template with placeholders (e.g., "added pattern %s to property %s")

# Advantages

  - Consistent naming: IDs follow predictable patterns based on hierarchy
  - Consistent messages: templates use standard grammar and structure
  - Extensible: add new change types by editing YAML, not code
  - Maintainable: single source of truth for all message definitions

# Status

This is an internal tool. The generated output (messages.yaml) can replace the
manually written messages in localizations_src. Additional work is needed to
ensure full coverage and handle translations.
*/
package generator
