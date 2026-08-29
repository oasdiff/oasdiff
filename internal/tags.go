package internal

import "slices"

// A tagDimension is one axis of a tag vocabulary: its values and how a value
// matches a row. Tag filtering is OR within a dimension and AND across
// dimensions: --tags request,response,add selects rows that are (request or
// response) and add. Values must be unique across the dimensions of one
// vocabulary so every tag names exactly one axis.
type tagDimension[T any] struct {
	values []string
	match  func(value string, row T) bool
}

func tagValues[T any](dimensions []tagDimension[T]) []string {
	var result []string
	for _, d := range dimensions {
		result = append(result, d.values...)
	}
	return result
}

func matchTagDimensions[T any](tags []string, dimensions []tagDimension[T], row T) bool {
	for _, d := range dimensions {
		matched, requested := false, false
		for _, tag := range tags {
			if !slices.Contains(d.values, tag) {
				continue
			}
			requested = true
			if d.match(tag, row) {
				matched = true
				break
			}
		}
		if requested && !matched {
			return false
		}
	}
	return true
}
