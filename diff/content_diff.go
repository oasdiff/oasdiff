package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ContentDiff describes the changes between content properties each containing media type objects: https://swagger.io/specification/#media-type-object
type ContentDiff struct {
	MediaTypeAdded    []string           `json:"added,omitempty" yaml:"added,omitempty"`
	MediaTypeDeleted  []string           `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	MediaTypeModified ModifiedMediaTypes `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// ModifiedMediaTypes is map of media type names to their respective diffs
type ModifiedMediaTypes map[string]*MediaTypeDiff

func newContentDiff() *ContentDiff {
	return &ContentDiff{
		MediaTypeAdded:    []string{},
		MediaTypeDeleted:  []string{},
		MediaTypeModified: ModifiedMediaTypes{},
	}
}

// Empty indicates whether a change was found in this element
func (diff *ContentDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.MediaTypeAdded) == 0 &&
		len(diff.MediaTypeDeleted) == 0 &&
		len(diff.MediaTypeModified) == 0
}

func getContentDiff(config *Config, state *state, content1, content2 openapi3.Content) (*ContentDiff, error) {
	diff, err := getContentDiffInternal(config, state, content1, content2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getContentDiffInternal(config *Config, state *state, content1, content2 openapi3.Content) (*ContentDiff, error) {

	initialResult := newContentDiff()
	processedDeleted := make(map[string]bool) // Keep track of deleted items that found an equivalent added item

	// 1. Diff media types with the same name. The rest are initial additions
	// and deletions, sorted, so that when several media types are equivalent the
	// same ones are paired on every run.
	added, deleted, common := keysDiff(content1, content2)
	initialResult.MediaTypeAdded = added
	initialResult.MediaTypeDeleted = deleted

	for _, name := range common {
		diff, err := getMediaTypeDiff(config, state, name, name, content1[name], content2[name])
		if err != nil {
			return nil, err
		}

		if !diff.Empty() {
			initialResult.MediaTypeModified[name] = diff
		}
	}

	// 2. Iteratively find equivalent pairs from initial Added/Deleted lists
	finalAdded := make([]string, 0, len(initialResult.MediaTypeAdded))
	finalModified := initialResult.MediaTypeModified // Start with exact matches

	for _, addedName := range initialResult.MediaTypeAdded {
		foundEquivalentDeleted := false
		for _, deletedName := range initialResult.MediaTypeDeleted {
			if processedDeleted[deletedName] { // Skip deleted items already paired
				continue
			}

			if isMediaTypeNamesEquivalent(addedName, deletedName) {
				// Found an equivalent pair: addedName <-> deletedName

				// Calculate diff
				diff, err := getMediaTypeDiff(config, state, deletedName, addedName, content1[deletedName], content2[addedName])
				if err != nil {
					return nil, err
				}

				// Add to modified, using the *new* name as the key seems more logical here,
				finalModified[addedName] = diff

				processedDeleted[deletedName] = true // Mark this deleted item as paired
				foundEquivalentDeleted = true
				break // Move to the next addedName
			}
		}

		if !foundEquivalentDeleted {
			// This addedName did not find an equivalent deletedName
			finalAdded = append(finalAdded, addedName)
		}
	}

	// 3. Collect final deletions (those not marked as processed)
	finalDeleted := make([]string, 0, len(initialResult.MediaTypeDeleted))
	for _, deletedName := range initialResult.MediaTypeDeleted {
		if !processedDeleted[deletedName] {
			finalDeleted = append(finalDeleted, deletedName)
		}
	}

	// 4. Construct final ContentDiff
	finalResult := &ContentDiff{
		MediaTypeAdded:    finalAdded,
		MediaTypeDeleted:  finalDeleted,
		MediaTypeModified: finalModified,
	}

	return finalResult, nil
}

func isMediaTypeNamesEquivalent(name1, name2 string) bool {
	// Check containment in both directions to see if they are related
	// This covers base -> specific refinement and specific -> base refinement
	// Although IsMediaTypeNameContained is designed for base->specific, checking both ways ensures we pair them.
	return IsMediaTypeNameContained(name1, name2) || IsMediaTypeNameContained(name2, name1)
}
