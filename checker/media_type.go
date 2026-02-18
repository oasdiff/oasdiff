package checker

// formatMediaTypeDetails returns media type context for messages.
// Returns empty string if only one media type exists to keep messages concise.
func formatMediaTypeDetails(mediaType string, totalMediaTypes int) string {
	if totalMediaTypes <= 1 {
		return ""
	}
	return "(media type: " + mediaType + ")"
}

// combineDetails combines multiple detail strings, filtering out empty ones.
func combineDetails(details ...string) string {
	var result string
	for _, d := range details {
		if d != "" {
			if result != "" {
				result += " "
			}
			result += d
		}
	}
	return result
}
