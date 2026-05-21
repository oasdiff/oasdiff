# Changelog

## Unreleased

### Fixed

- Suppressed false positive removed `anyOf`/`oneOf` subschema reports when an inline branch is refactored to a validation-equivalent `$ref`; mixed reports now list only branches that were actually removed.
