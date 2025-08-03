package internal

import (
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
	return getError(
		fmt.Errorf("failed to load %s spec from %s: %w", what, source.Out(), err),
		102,
	)
}

func getErrFailedToLoadSpecs(what string, path string, err error) *ReturnError {
	return getError(
		fmt.Errorf("failed to load %s specs from glob %q: %w", what, path, err),
		103,
	)
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

func getError(err error, code int) *ReturnError {
	return &ReturnError{err, code}
}
