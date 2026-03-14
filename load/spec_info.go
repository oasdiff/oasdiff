package load

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yargevad/filepathx"
)

// SpecInfo contains information about an OpenAPI spec and its metadata
type SpecInfo struct {
	Url     string
	Spec    *openapi3.T
	Version string
}

func (specInfo *SpecInfo) GetVersion() string {
	if specInfo == nil || specInfo.Version == "" {
		return "n/a"
	}
	return specInfo.Version
}

func newSpecInfo(spec *openapi3.T, path string) *SpecInfo {
	return &SpecInfo{
		Spec:    spec,
		Url:     path,
		Version: getVersion(spec),
	}
}

func getVersion(spec *openapi3.T) string {
	if spec == nil || spec.Info == nil {
		return ""
	}

	return spec.Info.Version
}

// NewSpecInfo creates a SpecInfo from a local file path, a URL, or stdin
func NewSpecInfo(loader *openapi3.Loader, source *Source, options ...Option) (*SpecInfo, error) {
	specInfo, err := loadSpecInfo(loader, source)
	if err != nil {
		return nil, err
	}
	specInfos := []*SpecInfo{specInfo}

	for _, option := range options {
		if specInfos, err = option(loader, specInfos); err != nil {
			return nil, err
		}
	}
	return specInfos[0], nil
}

// NewSpecInfoFromGlob creates SpecInfos from local files matching the specified glob parameter
func NewSpecInfoFromGlob(loader *openapi3.Loader, glob string, options ...Option) ([]*SpecInfo, error) {
	specInfos, err := fromGlob(loader, glob)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		if specInfos, err = option(loader, specInfos); err != nil {
			return nil, err
		}
	}
	return specInfos, nil
}

func loadSpecInfo(loader *openapi3.Loader, source *Source) (*SpecInfo, error) {
	s, err := from(loader, source)
	if err != nil {
		return nil, err
	}
	return newSpecInfo(s, source.DisplayPath()), nil
}

func fromGlob(loader *openapi3.Loader, glob string) ([]*SpecInfo, error) {
	files, err := filepathx.Glob(glob)
	if err != nil {
		return nil, err
	}
	result := make([]*SpecInfo, 0)
	for _, file := range files {
		spec, err := loader.LoadFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load %q: %w", file, err)
		}
		result = append(result, &SpecInfo{Url: file, Spec: spec})
	}

	if len(result) > 0 {
		return result, nil
	}

	if isUrl(glob) {
		return nil, errors.New("no matching files (should be a glob, not a URL)")
	}

	return nil, errors.New("no matching files")

}

func isUrl(spec string) bool {
	_, err := url.ParseRequestURI(spec)
	return err == nil
}
