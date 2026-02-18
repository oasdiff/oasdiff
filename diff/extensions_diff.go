package diff

// ExtensionsDiff describes the changes between a pair of sets of specification extensions: https://swagger.io/specification/#specification-extensions
type ExtensionsDiff InterfaceMapDiff

// Empty indicates whether a change was found in this element
func (diff *ExtensionsDiff) Empty() bool {
	return (*InterfaceMapDiff)(diff).Empty()
}

func getExtensionsDiff(config *Config, extensions1, extensions2 map[string]any) (*ExtensionsDiff, error) {
	if config.IsExcludeExtensions() {
		return nil, nil
	}

	// Filter out excluded extension names
	filtered1 := filterExtensions(extensions1, config)
	filtered2 := filterExtensions(extensions2, config)

	diff, err := getExtensionsDiffInternal(filtered1, filtered2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return (*ExtensionsDiff)(diff), nil
}

func getExtensionsDiffInternal(extensions1, extensions2 map[string]any) (*InterfaceMapDiff, error) {
	return getInterfaceMapDiff(extensions1, extensions2)
}

// filterExtensions returns a copy of the extensions map with excluded extensions removed
func filterExtensions(extensions map[string]any, config *Config) map[string]any {
	if len(config.ExcludeExtensions) == 0 {
		return extensions
	}

	filtered := make(map[string]any)
	for name, value := range extensions {
		if !config.IsExcludedExtension(name) {
			filtered[name] = value
		}
	}
	return filtered
}
