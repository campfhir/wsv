package record

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/campfhir/wsv/utils"
)

type RecordField struct {
	IsNull     bool
	Value      string
	FieldIndex int
	RowIndex   int
	FieldName  string
	IsHeader   bool
}

func (f *RecordField) CalculateFieldLength() int {
	v := f.SerializeText()
	return utf8.RuneCountInString(v)
}

func (f *RecordField) SerializeText() string {
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
