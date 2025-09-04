package internal

import (
	"fmt"
	"strings"
)

func ParseBool(str, format string) (bool, error) {
	a := strings.Split(format, "|")
	truth := "True"
	falsehood := "False"
	if len(a) >= 2 {
		truth = a[0]
		falsehood = a[1]
	}
	switch str {
	case truth:
		return true, nil
	case falsehood:
		return false, nil
	}
	return false, fmt.Errorf("could not parse '%s' as a bool", str)
}

// FormatBool returns the string representation of a bool in the requested format.
// Falls back to "True"/"False" if format not found.
func FormatBool(v bool, format string) string {
	a := strings.Split(format, "|")
	if len(a) < 2 {
		if v {
			return "True"
		}
		return "False"
	}
	if v {
		return a[0]
	}
	return a[1]
}
