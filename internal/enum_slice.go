package internal

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"slices"
	"strings"
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

// normalizeAllowedValues maps each value to its canonical allowed value,
// matching case-insensitively so a caller can pass any case and downstream
// consumers still receive the exact allowed string. Values with no
// case-insensitive match are reported as not allowed.
func (s *enumSliceValue) normalizeAllowedValues(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	var notAllowed []string
	for _, v := range values {
		if canonical, ok := s.canonicalValue(v); ok {
			normalized = append(normalized, canonical)
		} else {
			notAllowed = append(notAllowed, v)
		}
	}

	if len(notAllowed) > 0 {
		verb := "are"
		if len(notAllowed) == 1 {
			verb = "is"
		}
		return nil, fmt.Errorf("%s %s not one of the allowed values: %s", strings.Join(notAllowed, ","), verb, s.listOf())
	}

	return normalized, nil
}

// canonicalValue returns the allowed value equal to value (case-insensitive)
// and whether one was found.
func (s *enumSliceValue) canonicalValue(value string) (string, bool) {
	for _, allowed := range s.allowedValues {
		if strings.EqualFold(value, allowed) {
			return allowed, true
		}
	}
	return "", false
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

	value, err = s.normalizeAllowedValues(value)
	if err != nil {
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
