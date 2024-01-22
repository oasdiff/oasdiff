package delta

import (
	"github.com/tufin/oasdiff/diff"
)

func getParametersDelta(asymmetric bool, d *diff.ParametersDiffByLocation) WeightedDelta {
	if d.Empty() {
		return WeightedDelta{}
	}

	added := d.Added.Len()
	deleted := d.Deleted.Len()
	modified := d.Modified.Len()
	unchanged := d.Unchanged.Len()
	all := added + deleted + modified + unchanged

	// TODO: drill down into modified
	modifiedDelta := coefficient * float64(modified)

	return WeightedDelta{
		delta:  ratio(asymmetric, added, deleted, modifiedDelta, all),
		weight: all,
	}
}
