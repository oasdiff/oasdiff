package internal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
)

type ReturnError struct {
	error
	Code int
}

const generalExecutionErr = 100

func getErrInvalidFlags(err error) *ReturnError {
	return getError(
		err,
		101,
	)
}

func getErrFailedToLoadSpec(what string, source *load.Source, err error) *ReturnError {
	if flatErr := asFlattenError(err); flatErr != nil {
		return getErrFailedToFlatten(flatErr)
	}
	wrapped := fmt.Errorf("failed to load %s spec from %s: %w", what, source.Out(), err)
	if isExternalRefError(err) {
		return getErrDisallowedExternalRef(wrapped)
	}
	return getError(wrapped, 102)
}

func getErrFailedToLoadSpecs(what string, path string, err error) *ReturnError {
	if flatErr := asFlattenError(err); flatErr != nil {
		return getErrFailedToFlatten(flatErr)
	}
	wrapped := fmt.Errorf("failed to load %s specs from glob %q: %w", what, path, err)
	if isExternalRefError(err) {
		return getErrDisallowedExternalRef(wrapped)
	}
	return getError(wrapped, 103)
}

func getErrDiffFailed(err error) *ReturnError {
	return getError(
		fmt.Errorf("diff failed: %w", err),
		104,
	)
}

func getErrFailedPrint(what string, err error) *ReturnError {
	return getError(
		fmt.Errorf("failed to print %q: %w", what, err),
		105,
	)
}

func getErrFailedToLoadSeverityLevels(source string, err error) *ReturnError {
	return getError(
		fmt.Errorf("failed to load custom severity levels from %s: %w", source, err),
		106,
	)
}

func getErrConfigFileProblem(err error) *ReturnError {
	return getError(
		fmt.Errorf("failed to load config file: %w", err),
		107,
	)
}

func getErrUnsupportedFormat(format, cmd string) *ReturnError {
	return getError(
		fmt.Errorf("format %q is not supported by %q", format, cmd),
		110,
	)
}

func getErrTemplateNotSupported(format string) *ReturnError {
	supportedFormats := formatters.GetSupportedTemplateFormats()
	return getError(
		fmt.Errorf("template flag is not supported for format %q. Supported formats for templates are: %s", format, strings.Join(supportedFormats, ", ")),
		111,
	)
}

func getErrInvalidColorMode(err error) *ReturnError {
	return getError(
		err,
		114,
	)
}

func getErrCantProcessIgnoreFile(what string, err error) *ReturnError {
	return getError(
		fmt.Errorf("can't process %s ignore file: %w", what, err),
		121,
	)
}

// getErrFailedToFlatten returns the FlattenError unwrapped — its Error()
// already reports the offending file and the merge failure. The outer
// "failed to load spec" wrap is misleading because loading succeeded.
func getErrFailedToFlatten(err *load.FlattenError) *ReturnError {
	return getError(err, 122)
}

// getErrDisallowedExternalRef returns the dedicated exit code (123) for a spec
// that resolved an external $ref while --allow-external-refs=false. Callers
// (e.g. the GitHub Action) key off this code to surface a precise
// "set allow-external-refs" remedy without matching error text. 123 is kept
// under 125 to stay clear of the shell's reserved 126/127/128+ range.
func getErrDisallowedExternalRef(err error) *ReturnError {
	return getError(err, 123)
}

// asFlattenError returns the *load.FlattenError in err's chain, or nil if err is
// a genuine load failure. Centralised so every spec-loading site stays a single
// line at the call site.
func asFlattenError(err error) *load.FlattenError {
	if flatErr, ok := errors.AsType[*load.FlattenError](err); ok {
		return flatErr
	}
	return nil
}

// isExternalRefError reports whether err's chain contains a *load.ExternalRefError.
func isExternalRefError(err error) bool {
	var extErr *load.ExternalRefError
	return errors.As(err, &extErr)
}

func getError(err error, code int) *ReturnError {
	return &ReturnError{err, code}
}
