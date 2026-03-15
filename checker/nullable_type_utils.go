package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

// isNullTypeChange returns true if the only change in a type list is adding or removing "null"
func isNullTypeChange(typeDiff *diff.StringsDiff) bool {
	if typeDiff == nil {
		return false
	}
	return onlyNull(typeDiff.Added) && onlyNull(typeDiff.Deleted)
}

// nullAddedToTypeArray returns true if "null" was added to the type array (OpenAPI 3.1 nullable)
func nullAddedToTypeArray(typeDiff *diff.StringsDiff) bool {
	if typeDiff == nil {
		return false
	}
	for _, t := range typeDiff.Added {
		if t == "null" {
			return true
		}
	}
	return false
}

// nullRemovedFromTypeArray returns true if "null" was removed from the type array (OpenAPI 3.1 became not-nullable)
func nullRemovedFromTypeArray(typeDiff *diff.StringsDiff) bool {
	if typeDiff == nil {
		return false
	}
	for _, t := range typeDiff.Deleted {
		if t == "null" {
			return true
		}
	}
	return false
}

func onlyNull(types []string) bool {
	for _, t := range types {
		if t != "null" {
			return false
		}
	}
	return true
}
