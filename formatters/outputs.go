package formatters

type Output int

const (
	OutputDiff Output = iota
	OutputSummary
	OutputChangelog
	OutputChecks
	OutputExplain
	OutputFlatten
	OutputValidate
)
