package load

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/flatten/allof"
	"github.com/oasdiff/oasdiff/flatten/commonparams"
	"github.com/oasdiff/oasdiff/flatten/headers"
)

// Option functions can be used to preprocess specs after loading them
type Option func(*openapi3.Loader, []*SpecInfo) ([]*SpecInfo, error)

// WithIdentity returns the original SpecInfos
func WithIdentity() Option {
	return func(loader *openapi3.Loader, specInfos []*SpecInfo) ([]*SpecInfo, error) {
		return specInfos, nil
	}
}

// GetOption returns the requested option or the identity option
func GetOption(option Option, enable bool) Option {
	if !enable {
		return WithIdentity()
	}
	return option
}

// FlattenError reports a failure to merge allOf during WithFlattenAllOf.
// Returned wrapped so callers can use errors.As to distinguish a flatten
// failure (which happens after the spec has loaded successfully) from a
// genuine load failure.
type FlattenError struct {
	Url string
	Err error
}

func (e *FlattenError) Error() string {
	return fmt.Sprintf("failed to flatten allOf in %q: %s", e.Url, e.Err)
}

func (e *FlattenError) Unwrap() error { return e.Err }

// WithFlattenAllOf returns SpecInfos with flattened allOf
func WithFlattenAllOf() Option {
	return func(loader *openapi3.Loader, specInfos []*SpecInfo) ([]*SpecInfo, error) {
		var err error
		for _, specInfo := range specInfos {
			if specInfo.Spec, err = allof.MergeSpec(specInfo.Spec); err != nil {
				return nil, &FlattenError{Url: specInfo.Url, Err: err}
			}
		}
		return specInfos, nil
	}
}

// WithFlattenParams returns SpecInfos with Common Parameters combined into operation parameters
// See here for Common Parameters definition: https://swagger.io/docs/specification/describing-parameters/
func WithFlattenParams() Option {
	return func(loader *openapi3.Loader, specInfos []*SpecInfo) ([]*SpecInfo, error) {
		for _, specInfo := range specInfos {
			commonparams.Move(specInfo.Spec)
		}
		return specInfos, nil
	}
}

// WithLowercaseHeaders returns SpecInfos with header names converted to lowercase
func WithLowercaseHeaders() Option {
	return func(loader *openapi3.Loader, specInfos []*SpecInfo) ([]*SpecInfo, error) {
		for _, specInfo := range specInfos {
			headers.Lowercase(specInfo.Spec)
		}
		return specInfos, nil
	}
}
