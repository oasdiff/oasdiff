package validate_test

import (
	"errors"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/validate"
	"github.com/stretchr/testify/require"
)

func TestAllOfConflict(t *testing.T) {
	finding := validate.AllOfConflict(errors.New("unable to resolve Type conflict: all Type values must be identical"), "spec.yaml")

	require.Equal(t, validate.AllOfConflictID, finding.Id)
	require.Equal(t, checker.WARN, finding.Level)
	require.Equal(t, "spec.yaml", finding.Source.File)
	require.Contains(t, finding.Text, "unable to resolve Type conflict")
	require.NotEmpty(t, finding.Fingerprint)
}

// The merge reports which conflict it hit, not where, so the finding carries
// no line or column.
func TestAllOfConflict_HasNoLocation(t *testing.T) {
	finding := validate.AllOfConflict(errors.New("conflict"), "spec.yaml")

	require.Zero(t, finding.Source.Line)
	require.Zero(t, finding.Source.Column)
}

// The fingerprint distinguishes one conflict from another, so a spec whose
// conflict changes does not look like the same finding.
func TestAllOfConflict_FingerprintFollowsTheConflict(t *testing.T) {
	typeConflict := validate.AllOfConflict(errors.New("unable to resolve Type conflict"), "spec.yaml")
	enumConflict := validate.AllOfConflict(errors.New("unable to resolve Enum conflict"), "spec.yaml")

	require.NotEqual(t, typeConflict.Fingerprint, enumConflict.Fingerprint)
}
