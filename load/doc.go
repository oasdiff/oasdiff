/*
Package load provides functions to load OpenAPI specifications from various sources.

# Overview

The load package handles loading OpenAPI specs from files, URLs, stdin, and glob patterns.
It automatically resolves $ref references and extracts version information from the spec.

# Usage

Load a spec from a file, URL, or stdin:

	source, _ := load.NewSource("openapi.yaml")
	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), source)

Load multiple specs using glob patterns:

	specInfos, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "specs/*.yaml")

# Preprocessing Options

Options can preprocess specs after loading to improve diff accuracy:

	specInfo, err := load.NewSpecInfo(loader, source,
	    load.WithFlattenAllOf(),      // Merge allOf schemas into single schema
	    load.WithFlattenParams(),     // Move common path parameters to operations
	    load.WithLowercaseHeaders(),  // Normalize header names to lowercase
	)

These options use the flatten subpackages:
  - flatten/allof: merges allOf schemas for more accurate breaking change detection
  - flatten/commonparams: moves path-level parameters to operations for consistent comparison
  - flatten/headers: lowercases header names since HTTP headers are case-insensitive

# SpecInfo

SpecInfo wraps a loaded spec with metadata:
  - Spec: the parsed openapi3.T object with resolved references
  - Url: the source path/URL the spec was loaded from
  - Version: the API version extracted from info.version
*/
package load
