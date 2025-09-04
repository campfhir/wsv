package internal

import (
	"fmt"
	"strings"
)

func ParseBool(str string) (bool, error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "YES", "Yes", "yes", "Y", "y", "X", "x", "✅", "✓":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "NO", "No", "no", "N", "n", "", "❎":
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
