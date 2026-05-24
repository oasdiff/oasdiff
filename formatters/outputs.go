package formatters

type Output int

const (
	OutputDiff Output = iota
	OutputSummary
	OutputChangelog
	OutputChecks
	OutputFlatten
	OutputValidate
)
