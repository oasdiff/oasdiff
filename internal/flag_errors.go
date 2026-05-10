package internal

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// friendlyFlagError rewrites pflag's raw stdlib parse-error wrapping into a
// human-readable hint. For example,
//
//	invalid argument "x" for "--flatten-params" flag: strconv.ParseBool: parsing "x": invalid syntax
//
// becomes
//
//	invalid argument "x" for "--flatten-params" flag: must be true or false
//
// Errors that pflag did not produce are returned unchanged.
func friendlyFlagError(_ *cobra.Command, err error) error {
	var ive *pflag.InvalidValueError
	if !errors.As(err, &ive) {
		return err
	}

	hint := flagTypeHint(ive.GetFlag())
	if hint == "" {
		return err
	}

	flagName := "--" + ive.GetFlag().Name
	return fmt.Errorf("invalid argument %q for %q flag: %s", ive.GetValue(), flagName, hint)
}

// flagTypeHint returns a friendly description of the values a flag accepts,
// keyed off pflag's Value.Type() string. Returns "" when no specific hint
// applies, in which case the caller should leave the original error alone.
func flagTypeHint(f *pflag.Flag) string {
	if f == nil {
		return ""
	}
	switch f.Value.Type() {
	case "bool":
		return "must be true or false"
	case "int", "int8", "int16", "int32", "int64":
		return "must be an integer"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "must be a non-negative integer"
	case "float32", "float64":
		return "must be a number"
	case "duration":
		return "must be a duration like 30s or 5m"
	}
	return ""
}
