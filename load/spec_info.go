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
	// Sources holds the raw text of every file that contributed to Spec, keyed
	// by resolved path, when loaded via NewSpecInfoWithCapture; nil otherwise.
	Sources map[string]string
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

// NewSpecInfo creates a SpecInfo from a local file path, a URL, a git revision, or stdin
func NewSpecInfo(loader *openapi3.Loader, source *Source, options ...Option) (*SpecInfo, error) {
	return newSpecInfoFromSource(loader, source, false, options...)
}

// NewSpecInfoWithCapture is NewSpecInfo that also records the raw text of
// every file the loader reads (root and $ref'd) into SpecInfo.Sources, for
// slicing multi-file specs by origin file (the review bundle).
func NewSpecInfoWithCapture(loader *openapi3.Loader, source *Source, options ...Option) (*SpecInfo, error) {
	return newSpecInfoFromSource(loader, source, true, options...)
}

// newSpecInfoFromSource loads a single SpecInfo and applies the options. When
// capture is true it also records the raw text of every file read into
// SpecInfo.Sources. Capture is a constructor choice, not an Option, because the
// recorder must be installed before the load, whereas options run after it.
func newSpecInfoFromSource(loader *openapi3.Loader, source *Source, capture bool, options ...Option) (*SpecInfo, error) {
	var recorder *sourceCapture
	if capture {
		recorder = newSourceCapture()
	}
	specInfo, err := loadSpecInfo(loader, source, recorder)
	if err != nil {
		return nil, err
	}
	if recorder != nil {
		specInfo.Sources = recorder.asStrings()
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
	return newSpecInfoFromGlob(loader, glob, false, options...)
}

// NewSpecInfoFromGlobWithCapture is NewSpecInfoFromGlob that also records each
// spec's contributing file texts into its Sources (see NewSpecInfoWithCapture).
func NewSpecInfoFromGlobWithCapture(loader *openapi3.Loader, glob string, options ...Option) ([]*SpecInfo, error) {
	return newSpecInfoFromGlob(loader, glob, true, options...)
}

func newSpecInfoFromGlob(loader *openapi3.Loader, glob string, capture bool, options ...Option) ([]*SpecInfo, error) {
	specInfos, err := fromGlob(loader, glob, capture)
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

// NewSpecInfoFromData creates a SpecInfo from raw OpenAPI bytes already held in
// memory, labeling its source as name. name is loaded as the spec's path, so
// source-location reporting (e.g. the file name shown for each change) uses it
// rather than a temp path or an empty value. It is the in-memory counterpart to
// NewSpecInfo: no filesystem access, so relative file $refs are not resolved
// (intra-spec "#/..." refs still are). As elsewhere in this package, callers
// comparing two specs should use a fresh loader per spec so the loader's
// document cache does not collide when both share a name.
func NewSpecInfoFromData(loader *openapi3.Loader, data []byte, name string, options ...Option) (*SpecInfo, error) {
	spec, err := loader.LoadFromDataWithPath(data, &url.URL{Path: name})
	if err != nil {
		return nil, err
	}

	specInfos := []*SpecInfo{newSpecInfo(spec, name)}
	for _, option := range options {
		if specInfos, err = option(loader, specInfos); err != nil {
			return nil, err
		}
	}
	return specInfos[0], nil
}

func loadSpecInfo(loader *openapi3.Loader, source *Source, capture *sourceCapture) (*SpecInfo, error) {
	s, err := from(loader, source, capture)
	if err != nil {
		return nil, err
	}
	return newSpecInfo(s, source.DisplayPath()), nil
}

func fromGlob(loader *openapi3.Loader, glob string, capture bool) ([]*SpecInfo, error) {
	files, err := filepathx.Glob(glob)
	if err != nil {
		return nil, err
	}
	result := make([]*SpecInfo, 0)
	// Members load via LoadFromFile directly rather than through from():
	// glob expansion already proves each member is a local file, and
	// re-classifying the path as a Source could misread exotic names (a file
	// named "main:openapi.yaml" would parse as a git revision). Capture is
	// per member: each SpecInfo carries its own Sources, so a consumer can
	// slice each member's blocks from that member's files.
	for _, file := range files {
		l := loader
		var cap *sourceCapture
		if capture {
			cap = newSourceCapture()
			l = withCapture(loader, cap)
		}
		spec, err := l.LoadFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load %q: %w", file, err)
		}
		si := &SpecInfo{Url: file, Spec: spec}
		if cap != nil {
			si.Sources = cap.asStrings()
		}
		result = append(result, si)
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
