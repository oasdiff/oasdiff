package colorize

import "os"

var pipedOutput *bool

func SetPipedOutput(val *bool) *bool {
	save := pipedOutput
	pipedOutput = val
	return save
}

func isPipedOutput() bool {
	if pipedOutput != nil {
		return *pipedOutput
	}

	fi, _ := os.Stdout.Stat()
	a := (fi.Mode() & os.ModeCharDevice) == 0
	pipedOutput = &a
	return *pipedOutput
}
