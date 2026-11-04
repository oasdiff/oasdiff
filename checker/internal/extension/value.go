// Package extension reads the values of OpenAPI specification extensions.
package extension

import (
	"encoding/json"
	"fmt"
)

// ValueText writes an extension value as it appears in the spec, since %v
// prints a decoded object as a Go map.
func ValueText(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	text, err := json.Marshal(value)
	// a value set in code, rather than decoded from the spec, may not marshal
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(text)
}
