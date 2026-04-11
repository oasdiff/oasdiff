package diff_test

import "fmt"

// getPrefixFile returns the path to a file under data/prefix/.
func getPrefixFile(file string) string {
	return fmt.Sprintf("../data/prefix/%s", file)
}

// getXOfFile returns the path to a file under data/x-of/.
func getXOfFile(file string) string {
	return fmt.Sprintf("../data/x-of/%s", file)
}

// getXOfTitlesFile returns the path to a file under data/x-of-titles/.
func getXOfTitlesFile(file string) string {
	return fmt.Sprintf("../data/x-of-titles/%s", file)
}
