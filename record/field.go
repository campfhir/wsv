package record

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/campfhir/wsv/utils"
)

type Field struct {
	IsNull     bool
	Value      string
	FieldIndex int
	RowIndex   int
	FieldName  string
	IsHeader   bool
}

// Computes the rune length of the serialized value
func (f *Field) CalculateFieldLength() int {
	v := f.SerializeText()
	return utf8.RuneCountInString(v)
}

// Serializes the values of the field
//
// - escaping whitespaces, double quoutes, and hyphens from the records value
//
// - returns the literal `-` character for null
//
// - `""` for an empty string
func (f *Field) SerializeText() string {
	wrapped := false
	if f.IsNull {
		return "-"
	}
	v := f.Value

	v = strings.ReplaceAll(v, `"`, `""`)
	if strings.Contains(v, `""`) && !wrapped {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	if strings.Contains(v, "-") {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	v = strings.ReplaceAll(v, "\n", `"/"`)
	if strings.Contains(v, `"/"`) && !wrapped {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	if strings.ContainsFunc(v, utils.IsFieldDelimiter) && !wrapped {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	if v == "" {
		v = `""`
	}
	return v
}
