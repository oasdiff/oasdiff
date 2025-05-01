package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

// ContentDiff describes the changes between content properties each containing media type objects: https://swagger.io/specification/#media-type-object
type ContentDiff struct {
	MediaTypeAdded    utils.StringList   `json:"added,omitempty" yaml:"added,omitempty"`
	MediaTypeDeleted  utils.StringList   `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	MediaTypeModified ModifiedMediaTypes `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// ModifiedMediaTypes is map of media type names to their respective diffs
type ModifiedMediaTypes map[string]*MediaTypeDiff

func newContentDiff() *ContentDiff {
	return &ContentDiff{
		MediaTypeAdded:    utils.StringList{},
		MediaTypeDeleted:  utils.StringList{},
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

	// 1. Find exact matches and initial deletions
	for name1, media1 := range content1 {

		if media2, ok := content2[name1]; ok { // Exact match found

			diff, err := getMediaTypeDiff(config, state, name1, name1, media1, media2)
			if err != nil {
				return nil, err
			}

			if !diff.Empty() {
				initialResult.MediaTypeModified[name1] = diff
			}
			// Mark as processed (implicitly, by not adding to deleted)
		} else {
			// No exact match, potential deletion
			initialResult.MediaTypeDeleted = append(initialResult.MediaTypeDeleted, name1)
		}
	}

	// 2. Find initial additions
	for name2 := range content2 {
		if _, ok := content1[name2]; !ok {
			initialResult.MediaTypeAdded = append(initialResult.MediaTypeAdded, name2)
		}
	}

	// 3. Iteratively find equivalent pairs from initial Added/Deleted lists
	finalAdded := make(utils.StringList, 0, len(initialResult.MediaTypeAdded))
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

	// 4. Collect final deletions (those not marked as processed)
	finalDeleted := make(utils.StringList, 0, len(initialResult.MediaTypeDeleted))
	for _, deletedName := range initialResult.MediaTypeDeleted {
		if !processedDeleted[deletedName] {
			finalDeleted = append(finalDeleted, deletedName)
		}
	}

	// 5. Construct final ContentDiff
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
