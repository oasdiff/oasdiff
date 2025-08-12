package diff

import (
	"github.com/oasdiff/oasdiff/utils"
)

// PathParamsMap handles path param renaming
// for example:
// person/{personName} -> /person/{name}
// in such cases, PathParamsMap stores the param mapping:
// personName -> name
type PathParamsMap utils.StringMap

func NewPathParamsMap(pathParams1, pathParams2 []string) (PathParamsMap, bool) {
	len1 := len(pathParams1)

	if len1 != len(pathParams2) {
		return nil, false
	}

	result := make(PathParamsMap, len1)

	for i, pathParam1 := range pathParams1 {
		result[pathParam1] = pathParams2[i]
	}
	return result, true
}

func (pathParamsMap PathParamsMap) find(pathParam1, pathParam2 string) bool {
	if len(pathParamsMap) == 0 {
		return pathParam1 == pathParam2
	}
	return pathParamsMap[pathParam1] == pathParam2
}
