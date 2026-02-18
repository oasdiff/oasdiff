package internal

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/utils"
)

// enumSliceValue is like stringSliceValue with allowed values
type enumSliceValue struct {
	value         *[]string
	allowedValues []string
	changed       bool
}

func newEnumSliceValue(allowedValues []string, val []string) *enumSliceValue {
	result := new(enumSliceValue)
	slices.Sort(allowedValues)
	result.allowedValues = allowedValues
	result.value = &val
	return result
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)

	records, err := csvReader.Read()
	if err != nil {
		return []string{}, err
	}

	return records, nil
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (s *enumSliceValue) checkAllowedValues(values []string) error {
	if notAllowed := utils.StringSetFromSlice(values).Minus(utils.StringSetFromSlice(s.allowedValues)); !notAllowed.Empty() {
		verb := "are"
		if len(notAllowed) == 1 {
			verb = "is"
		}
		return fmt.Errorf("%s %s not one of the allowed values: %s", strings.Join(notAllowed.ToStringList(), ","), verb, s.listOf())
	}
	return nil
}

func (s *enumSliceValue) listOf() string {
	l := len(s.allowedValues)
	switch l {
	case 0:
		return "no options available"
	case 1:
		return s.allowedValues[0]
	case 2:
		return s.allowedValues[0] + " or " + s.allowedValues[1]
	default:
		return strings.Join(s.allowedValues[:l-1], ", ") + ", or " + s.allowedValues[l-1]
	}
}

func (s *enumSliceValue) Set(val string) error {
	value, err := readAsCSV(val)
	if err != nil {
		return err
	}

	if err := s.checkAllowedValues(value); err != nil {
		return err
	}

	if !s.changed {
		*s.value = value
	} else {
		*s.value = append(*s.value, value...)
	}
	s.changed = true
	return nil
}

func (s *enumSliceValue) Type() string {
	return "strings"
}

func (s *enumSliceValue) String() string {
	str, _ := writeAsCSV(*s.value)
	return str
}
