package reader

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	doc "github.com/campfhir/wsv/document"

	"github.com/campfhir/wsv/utils"
)

func toPointer[t any](a t) *t {
	return &a
}

func TestQuotedValuesLine(t *testing.T) {
	line := `"Given Name" "Family Name" "Date of Birth" "Favorite Color"`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 4 {
		t.Errorf("Parse line expected 4 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "Given Name" {
		t.Errorf(`expected field 1 to be Given Name but got %+v\n`, r[0])
	}
	if r[1].IsNull || r[1].Value != "Family Name" {
		t.Errorf(`expected field 2 to be Family Name but got %+v\n`, r[1])
	}
	if r[2].IsNull || r[2].Value != "Date of Birth" {
		t.Errorf(`expected field 3 to be Date of Birth but got %+v\n`, r[2])
	}
	if r[3].IsNull || r[3].Value != "Favorite Color" {
		t.Errorf(`expected field 3 to be Favorite Color but got %+v\n`, r[2])
	}
}

func TestIsEmptyStringLiteral(t *testing.T) {
	t1 := []byte{'\t', '"', '"', '\t'}
	t1_p := []*byte{}
	for _, i := range t1 {
		t1_p = append(t1_p, toPointer(i))
	}

	t2 := []byte{'"', '"', '\t'}
	t2_p := []*byte{nil}
	for _, i := range t2 {
		t2_p = append(t2_p, toPointer(i))
	}

	t3 := []byte{'"', '"'}
	t3_p := []*byte{nil}
	for _, i := range t3 {
		t3_p = append(t3_p, toPointer(i))
	}
	t3_p = append(t3_p, nil)

	t4 := []byte{' ', '"', '"', ' '}
	t4_p := []*byte{}
	for _, i := range t4 {
		t4_p = append(t4_p, toPointer(i))
	}

	t5 := []byte{'\t', '"', '"'}
	t5_p := []*byte{}
	for _, i := range t5 {
		t5_p = append(t5_p, toPointer(i))
	}
	t5_p = append(t5_p, nil)

	if !utils.IsLiteralEmptyString(t1_p) {
		t.Error("failed test 1")
	}
	if !utils.IsLiteralEmptyString(t2_p) {
		t.Error("failed test 2")
	}
	if !utils.IsLiteralEmptyString(t3_p) {
		t.Error("failed test 3")
	}
	if !utils.IsLiteralEmptyString(t4_p) {
		t.Error("failed test 4")
	}
	if !utils.IsLiteralEmptyString(t5_p) {
		t.Error("failed test 5")
	}
}

func TestEmptyStringLiteral(t *testing.T) {
	line := `India						""					🇮🇳			  -`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 4 {
		t.Errorf("Parse line expected 4 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "India" {
		t.Errorf(`expected field 1 to be India but got %+v\n`, r[0])
	}
	if r[1].IsNull || r[1].Value != "" {
		t.Errorf(`expected field 2 to be [EMPTY STRING] but got %+v\n`, r[1])
	}
	if r[2].IsNull || r[2].Value != "🇮🇳" {
		t.Errorf(`expected field 3 to be 🇮🇳 but got %+v\n`, r[2])
	}
	if !r[3].IsNull || r[3].Value != "" {
		t.Errorf(`expected field 3 to be [NULL] but got %+v\n`, r[2])
	}
}

func TestReadColumnAndDataWithSpaces(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"Given Name" "Family Name" "Date of Birth" "Favorite Color"`)
	lines = append(lines, `"Jean Smith" "Le Croix" "Jan 01 2023" "Space Purple"`)
	lines = append(lines, `"Mary Jane" "Vasquez Rojas" "Feb 02 2021" "Midnight Grey"`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := NewReader(str)
	line, err := r.Read()

	if err != nil {
		t.Error(err)
	}
	if line.FieldCount() < 4 {
		t.Errorf("expected 4 fields but only got %d instead", line.FieldCount())
	}
	if field, err := line.Field(0); err != nil || !field.IsHeader || field.Value != "Given Name" {
		t.Errorf("Header 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || !field.IsHeader || field.Value != "Family Name" {
		t.Errorf("Header 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || !field.IsHeader || field.Value != "Date of Birth" {
		t.Errorf("Header 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsHeader || field.Value != "Favorite Color" {
		t.Errorf("Header 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "Given Name" || field.Value != "Jean Smith" {
		t.Errorf("Row 1 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "Family Name" || field.Value != "Le Croix" {
		t.Errorf("Row 1 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "Date of Birth" || field.Value != "Jan 01 2023" {
		t.Errorf("Row 1 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "Favorite Color" || field.Value != "Space Purple" {
		t.Errorf("Row 1 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "Given Name" || field.Value != "Mary Jane" {
		t.Errorf("Row 2 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "Family Name" || field.Value != "Vasquez Rojas" {
		t.Errorf("Row 2 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "Date of Birth" || field.Value != "Feb 02 2021" {
		t.Errorf("Row 2 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "Favorite Color" || field.Value != "Midnight Grey" {
		t.Errorf("Row 2 Column 4 does not match %+v %s", field, err)
	}
}

func TestParseDataWithEmoji(t *testing.T) {
	fields, err := parseLine(1, []byte("France						Paris	            🇫🇷			  \"The Eiffel Tower was built for the 1889 World's Fair.\"/\"It was almost torn down afterwards.\""))
	if err != nil {
		t.Error(err)
		return

	}
	if len(fields) != 4 {
		t.Errorf("expected 4 fields but only got %d instead", len(fields))
		return
	}

	if fields[0].IsNull || fields[0].Value != "France" || fields[0].IsComment {
		t.Error("expected field 1 to be [France] but got", fields[0].Value, "instead")
	}

	if fields[1].IsNull || fields[1].Value != "Paris" || fields[1].IsComment {
		t.Error("expected field 2 to be [Paris] buParist got", fields[1].Value, "instead")
	}

	if fields[2].IsNull || fields[2].Value != "🇫🇷" || fields[2].IsComment {
		t.Error("expected field 3 to be [🇫🇷] but got", fields[2].Value, "instead")
	}

	if fields[3].IsNull || fields[3].Value != "The Eiffel Tower was built for the 1889 World's Fair."+string('\n')+"It was almost torn down afterwards." || fields[3].IsComment {
		t.Error("expected field 4 to be [The Eiffel Tower was built for the 1889 World's Fair.\\nIt was almost torn down afterwards.] but got", fields[3].Value, "instead")
	}
}

func TestReadColumnAndDataWithoutSpaces(t *testing.T) {
	lines := make([]string, 0, 3)
	lines = append(lines, `fName lName dob gender`)
	lines = append(lines, `john doe 2023-01-01 M`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || !field.IsHeader || field.Value != "fName" {
		t.Errorf("Header 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || !field.IsHeader || field.Value != "lName" {
		t.Errorf("Header 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || !field.IsHeader || field.Value != "dob" {
		t.Errorf("Header 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsHeader || field.Value != "gender" {
		t.Errorf("Header 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "fName" || field.Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "lName" || field.Value != "doe" {
		t.Errorf("Row 1 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "dob" || field.Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "gender" || field.Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "fName" || field.Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "lName" || field.Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "dob" || field.Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "gender" || field.Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v %s", field, err)
	}
}

func TestReadColumnAndDataWithComments(t *testing.T) {
	lines := make([]string, 0, 4)
	lines = append(lines, `fName lName dob gender #these are headers`)
	lines = append(lines, `john doe 2023-01-01 M #this data is probably not accurate`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	lines = append(lines, `#this is a comment`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if !line.IsHeaderLine() {
		t.Error("expected line", line.LineNumber(), "to be designated as a header but was not marked as such")
	}
	if field, err := line.Field(0); err != nil || !field.IsHeader || field.Value != "fName" {
		t.Errorf("Header 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || !field.IsHeader || field.Value != "lName" {
		t.Errorf("Header 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || !field.IsHeader || field.Value != "dob" {
		t.Errorf("Header 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsHeader || field.Value != "gender" {
		t.Errorf("Header 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if line.Comment() != "this data is probably not accurate" {
		t.Errorf("Expected row to have a comment but got %s", err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "fName" || field.Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "lName" || field.Value != "doe" {
		t.Errorf("Row 1 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "dob" || field.Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "gender" || field.Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "fName" || field.Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "lName" || field.Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "dob" || field.Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "gender" || field.Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if line.FieldCount() != 0 && line.Comment() == "this is a comment" {
		t.Error("expected the line to have a length of 0 but got", line.FieldCount(), "instead")
		return
	}

	if len(line.Comment()) == 0 {
		t.Error("expected this line to contain a comment but got", line.Comment(), "instead")
	}
}

func TestParseWithEmptyLinesAndComments(t *testing.T) {
	lines := make([]string, 0, 4)
	lines = append(lines, `fName lName dob gender #these are headers`)
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "#here we go!")
	lines = append(lines, `john doe 2023-01-01 M #this data is probably not accurate`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	lines = append(lines, `#this is a comment`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if !line.IsHeaderLine() {
		t.Error("expected line", line.LineNumber(), "to be designated as a header but was not marked as such")
	}
	if field, err := line.Field(0); err != nil || !field.IsHeader || field.Value != "fName" {
		t.Errorf("Header 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || !field.IsHeader || field.Value != "lName" {
		t.Errorf("Header 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || !field.IsHeader || field.Value != "dob" {
		t.Errorf("Header 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsHeader || field.Value != "gender" {
		t.Errorf("Header 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if line.IsHeaderLine() {
		t.Error("expected line", line.LineNumber(), "to be NOT designated as a header but was not marked as such")
	}
	if err != nil {
		t.Error(err)
	}
	if line.FieldCount() != 0 {
		t.Error("expected row 2 to have zero fields but got", line.FieldCount(), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.FieldCount() != 0 {
		t.Error("expected row 3 to have zero fields but got", line.FieldCount(), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.FieldCount() > 0 {
		t.Error("expected row 4 to have 0 data field and 1 comment but got", line.FieldCount(), "instead")
		return
	}
	if len(line.Comment()) == 0 {
		t.Error("expected a comment but got", line.Comment(), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if line.Comment() != "this data is probably not accurate" {
		t.Errorf("Expected row to have a comment but got %+v", err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "fName" || field.Value != "john" {
		t.Errorf("Row 5 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "lName" || field.Value != "doe" {
		t.Errorf("Row 5 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "dob" || field.Value != "2023-01-01" {
		t.Errorf("Row 5 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "gender" || field.Value != "M" {
		t.Errorf("Row 5 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != "fName" || field.Value != "jane" {
		t.Errorf("Row 6 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != "lName" || field.Value != "smith" {
		t.Errorf("Row 6 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != "dob" || field.Value != "2024-01-01" {
		t.Errorf("Row 6 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != "gender" || field.Value != "F" {
		t.Errorf("Row 6 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if line.FieldCount() != 0 {
		t.Error("expected row 7 to have a length of 1 but got", line.FieldCount(), "instead")
		return
	}

	if len(line.Comment()) == 0 {
		t.Error("expected row 7 to contain a comment but got", line.Comment(), "instead")
	}
	if r.CurrentRow() != 7 {
		t.Error("expected current row to be 7 but got", r.CurrentRow(), "instead")
	}
}

func TestParseLineWithComments(t *testing.T) {
	line := `john "hungry ""doe""" #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLine(t *testing.T) {
	line := `john "hungry"/"doe" #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != "hungry\ndoe" {
		t.Errorf("expected field 2 to be hungry|doe but got %+v instead", r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithTrailingNewLineFollowedBySpace(t *testing.T) {
	line := `john "hungry"/"doe"/" " #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != `hungry`+string('\n')+`doe`+string('\n')+` ` {
		t.Errorf(`expected field 2 to be [hungry"\\n"doe\\n ] but got [%s] instead`, r[1].Value)
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLineAndEscapedDoubleQuotes(t *testing.T) {
	line := `john "hungry"""/"""doe" #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != `hungry"`+string('\n')+`"doe` {
		t.Errorf(`expected field 2 to be hungry"\\n"doe but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestQuotedValuesStartWith(t *testing.T) {
	line := `"john" "hungry" "hippo"`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "john" {
		t.Errorf(`expected field 1 to be john but got %+v\n`, r[0])
	}
	if r[1].IsNull || r[1].Value != "hungry" {
		t.Errorf(`expected field 2 to be hungry but got %+v\n`, r[1])
	}
	if r[2].IsNull || r[2].Value != "hippo" {
		t.Errorf(`expected field 3 to be hippo but got %+v\n`, r[2])
	}
}

func TestParseLineNullEndValue(t *testing.T) {
	line := `john hungry	-`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	// t.Errorf("%+v", r)
	if r[0].IsNull || r[0].Value != "john" {
		t.Error("expected field 1 to be john but got", r[0])
	}
	if r[1].IsNull || r[1].Value != "hungry" {
		t.Errorf(`expected field 2 to be hungry but got %+v\n`, r[1])
	}
	if !r[2].IsNull || r[2].Value != "" {
		t.Errorf(`expected field 3 to be null but got %+v\n`, r[2])
	}
}

func TestLiterallyDashValue(t *testing.T) {
	line := `"john-smith" hungry hippo`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "john-smith" {
		t.Error("expected field 1 to be john but got", r[0])
	}
	if r[1].IsNull || r[1].Value != "hungry" {
		t.Errorf(`expected field 2 to be hungry but got %+v\n`, r[1])
	}
	if r[1].IsNull || r[2].Value != `hippo` {
		t.Error(`expected field 3 to be hippo but got`, r[2].Value)
	}
}

func TestParseInvalidDashValue(t *testing.T) {
	line := `john -hungry	"hippo"`
	_, err := parseLine(1, []byte(line))
	if err.Error() != "john -hungry	\"hippo\"\n"+stringPadLeft("^", utf8.RuneCountInString("john "))+"\nparse error on line 1, column 5: null `-` specifier cannot be included without white space surrounding, unless it is the last value in the line. To record a literal `-` please wrap the value in double quotes" {
		t.Errorf("\n%v\n", err)
	}
}

func TestParseInvalidDashInLastField(t *testing.T) {
	line := `john hungry -hippo`
	_, err := parseLine(1, []byte(line))
	if err == nil {
		t.Error("expected this to throw an error")
	}
}

func TestParseBareDoubleQuote(t *testing.T) {
	line := `India						"""					🇮🇳			  -`
	r, err := parseLine(1, []byte(line))
	if err == nil {
		t.Error("Should have returned a bare double quote error", r)
		return
	}
}

func TestParseBareDoubleQuoteLastChar(t *testing.T) {
	line := `India						"""`
	r, err := parseLine(1, []byte(line))
	if err == nil {
		t.Error("Should have returned a bare double quote error", "["+r[1].Value+"]")
		return
	}
}

func TestParseLineNullMiddleValue(t *testing.T) {
	line := `john -	hippo`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	// t.Errorf("%+v", r)
	if r[0].IsNull || r[0].Value != "john" {
		t.Error("expected field 1 to be john but got", r[0])
	}
	if !r[1].IsNull || r[1].Value != "" {
		t.Errorf(`expected field 2 to be null but got %+v\n`, r[1])
	}
	if r[2].Value != `hippo` {
		t.Error(`expected field 3 to be hippo but got`, r[2].Value)
	}
}

func TestParseLineNullValue(t *testing.T) {
	line := `- hungry	hippo`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if !r[0].IsNull || r[0].Value != "" {
		t.Error("expected field 1 to be null but got", r[0])
	}
	if r[1].Value != `hungry` {
		t.Error(`expected field 2 to be hungry but got`, r[1].Value)
	}
	if r[2].Value != `hippo` {
		t.Error(`expected field 3 to be hippo but got`, r[2].Value)
	}
}

func TestParseTabs(t *testing.T) {
	line := `john	"hungry ""doe"""	#this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseWithHashSignInData(t *testing.T) {
	line := `john "hungry# ""doe"""`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].Value != "john" {
		t.Error("Expect john but got", r[0].Value)
	}
	if r[1].Value != `hungry# "doe"` {
		t.Error(`Expect hungry "doe" but got`, r[1].Value)
	}
}

func TestParseLineWithDoubleQuotes(t *testing.T) {
	line := `john "hungry ""doe"""`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].Value != "john" {
		t.Error("Expect john but got", r[0].Value)
	}
	if r[1].Value != `hungry "doe"` {
		t.Error(`Expect hungry "doe" but got`, r[1].Value)
	}
}

func TestParseLineWithStartingDoubleQuotes(t *testing.T) {
	line := `john """hungry"""`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].Value != "john" {
		t.Error("Expect john but got", r[0].Value)
	}
	if r[1].Value != `"hungry"` {
		t.Error(`Expect "hungry" but got`, r)
	}
}

func TestParseLineWithoutDoubleQuotes(t *testing.T) {
	line := `john hungry doe`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
	}
}

func TestParseInvalidLine(t *testing.T) {
	line := `john hungry "doe`
	_, err := parseLine(1, []byte(line))
	if err.Error() != "john hungry \"doe\n"+stringPadLeft("^", utf8.RuneCountInString(`john hungry `))+"\n"+`parse error on line 1, column 12: bare " in non-quoted-field` {
		t.Errorf("\n%v\n", err)
	}
}

func TestParseComplexDoubleQuoteLine(t *testing.T) {
	line := `"""fName""s" """lName""" """dob""" """gender"""`
	fields, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(fields) != 4 {
		t.Errorf("Parse line expected 4 record but got %+v (%d)", fields, len(fields))
		return
	}
}

func TestParseEmptyComment(t *testing.T) {
	line := `john hungry #`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
	}
}

func TestParseCommentWithJustSpaces(t *testing.T) {
	line := `john hungry #      `
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
	}
}

func TestReadColumnAndDataWithDoubleQuotes(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"""fName""s" """lName""" """dob""" """gender"""`)
	lines = append(lines, `john "hungry ""doe""" 2023-01-01 M`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}

	if field, err := line.Field(0); err != nil || !field.IsHeader || field.Value != `"fName"s` {
		t.Errorf("Header 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || !field.IsHeader || field.Value != `"lName"` {
		t.Errorf("Header 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || !field.IsHeader || field.Value != `"dob"` {
		t.Errorf("Header 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsHeader || field.Value != `"gender"` {
		t.Errorf("Header 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != `"fName"s` || field.Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != `"lName"` || field.Value != `hungry "doe"` {
		t.Errorf("Row 1 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != `"dob"` || field.Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != `"gender"` || field.Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != `"fName"s` || field.Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != `"lName"` || field.Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != `"dob"` || field.Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || field.IsHeader || field.FieldName != `"gender"` || field.Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v %s", field, err)
	}
}

func TestNullRemainingColumns(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"""fName""s" """lName""" """dob""" """gender"""`)
	lines = append(lines, `john "hungry ""doe""" 2023-01-01`)
	lines = append(lines, `jane smith 2024-01-01`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := NewReader(str)
	r.NullTrailingColumns = true
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.Value != `"fName"s` || !field.IsHeader {
		t.Errorf("Header 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.Value != `"lName"` || !field.IsHeader {
		t.Errorf("Header 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.Value != `"dob"` || !field.IsHeader {
		t.Errorf("Header 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsHeader || field.Value != `"gender"` {
		t.Errorf("Header 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.Value != "john" || field.FieldName != `"fName"s` || field.IsHeader {
		t.Errorf("Row 1 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != `"lName"` || field.Value != `hungry "doe"` {
		t.Errorf("Row 1 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != `"dob"` || field.Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsNull || field.FieldName != `"gender"` || field.Value != "" {
		t.Errorf("Row 1 Column 4 does not match %+v %s", field, err)
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if field, err := line.Field(0); err != nil || field.IsHeader || field.FieldName != `"fName"s` || field.Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v %s", field, err)
	}
	if field, err := line.Field(1); err != nil || field.IsHeader || field.FieldName != `"lName"` || field.Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v %s", field, err)
	}
	if field, err := line.Field(2); err != nil || field.IsHeader || field.FieldName != `"dob"` || field.Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v %s", field, err)
	}
	if field, err := line.Field(3); err != nil || !field.IsNull || field.FieldName != `"gender"` || field.Value != "" {
		t.Errorf("Row 2 Column 4 does not match %+v %s", field, err)
	}
}

func TestParseLineStartsWithNewLineAndEscapedDoubleQuotes(t *testing.T) {
	line := `john ""/"hungry"""/"""doe"/"" #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {

		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != string('\n')+`hungry"`+string('\n')+`"doe`+string('\n') {
		t.Errorf(`expected field 2 to be \\nhungry"\\n"doe\\n but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLineAtEndOfValueAndEscapedDoubleQuotes(t *testing.T) {
	line := `john "hungry"""/"""doe"/"" #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {

		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != `hungry"`+string('\n')+`"doe`+string('\n') {
		t.Errorf(`expected field 2 to be hungry"\\n"doe\\n but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLineAtEndOfValueFollowedByASpaceAndEscapedDoubleQuotes(t *testing.T) {
	line := `john " hungry"""/"""doe"/" " #this is a valid comment`
	r, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != ` hungry"`+string('\n')+`"doe`+string('\n')+` ` {
		t.Errorf(`expected field 2 to be hungry"\\n"doe\\n but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestNewLineParsing(t *testing.T) {
	line := `"john"       "doe"       "favorite colors:"/"red"/"blue"/"green"`
	r, err := parseLine(1, []byte(line))
	exp := "favorite colors:" + string('\n') + "red" + string('\n') + "blue" + string('\n') + "green"
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "john" {
		t.Errorf(`expected field 2 to be [john] but got %+v instead`, r[1])
	}
	if r[1].IsNull || r[1].Value != "doe" {
		t.Errorf(`expected field 2 to be [doe] but got %+v instead`, r[1])
	}
	if r[2].IsNull || r[2].Value != exp {
		t.Errorf("Parse line expected the 3 field to be a [%s] but got [%s] instead", exp, r[2].Value)
	}
}

func TestReadDataWithBareQuote(t *testing.T) {
	file, err := os.Open("testdata/bare-quote.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	_, err = r.Read()
	if err == nil {
		t.Error("Should have failed")
	}
	if err.Error() != "\"First name\" \"This is a date # line is missing ending quote\n            ^\nparse error on line 1, column 13: bare \" in non-quoted-field" {
		t.Error(err)

	}
}

func TestReadDataMissingData(t *testing.T) {
	file, err := os.Open("testdata/missing-many-fields-tabular-records.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	_, err = r.ReadAll()
	if err == nil {
		t.Error("Should have failed")
		return
	}
	if err.Error() != "record on line 2: does not have the correct field count, has 2/4, the values for the headers (\"Date of Birth\"	\"Favorite Color\") are missing\nrecord on line 3: does not have the correct field count, has 3/4, the value for the header (\"Favorite Color\") is missing" {
		t.Error(err)
	}
}

func TestReadExtraRecords(t *testing.T) {
	file, err := os.Open("testdata/extra-tabular-records.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	l, err := r.ReadAll()
	if err == nil {
		t.Errorf("%d/%d\n", len(r.Headers()), l[2].FieldCount())
		t.Error("Should have failed")
		return
	}
	if err.Error() != "record on line 3: does not have the correct field count, has 5/4, the value (\"Dark Mode\") is extra\nrecord on line 4: does not have the correct field count, has 6/4, the values (\"Light Mode\", Admin) are extra" {
		t.Error(err)
	}
}

func TestReadDataWithNewLineInValue(t *testing.T) {
	file, err := os.Open("testdata/data-with-new-line-in-value.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 3 {
		t.Error("expected to have 3 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "first name" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "last name" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "comment" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "john" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "doe" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "favorite colors:\nred\nblue\ngreen" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "doe" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "favorite colors:\ngreen\nyellow" {
		t.Error(err)
		return
	}
}

func TestReadDataWithSpaces(t *testing.T) {
	file, err := os.Open("testdata/data-with-spaces.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Given Name" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "Family Name" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Date of Birth" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "Favorite Color" {
		t.Error("expect header 4 to be [Favorite Color] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Jean Smith" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "Le Croix" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Jan 01 2023" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "Space Purple" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [Space Purple] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Mary Jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "Vasquez Rojas" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Feb 02 2021" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "Midnight Grey" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [Midnight Grey] but got", field.Value, "instead")
	}
}

func TestReadDataWithoutSpace(t *testing.T) {
	file, err := os.Open("testdata/data-without-spaces.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "fName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "lName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "dob" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "john" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "doe" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2023-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "smith" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2024-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "F" {
		t.Error("expect row 3, field 4 to be [F] but got", field.Value, "instead")
	}
}

func TestReadOmittedColumnsToNull(t *testing.T) {
	file, err := os.Open("testdata/omitted-columns-to-null.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	r.NullTrailingColumns = true
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "fName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "lName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "dob" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "john" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "doe" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2023-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [NULL] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "smith" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2024-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [NULL] but got", field.Value, "instead")
	}
}

func TestReadSimpleWithTabs(t *testing.T) {
	file, err := os.Open("testdata/simple-with-tabs.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 3 {
		t.Error("expected to have 3 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Name" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "Age" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Color" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Scott" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "21" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Red" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Josh" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "18" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Blue" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "34" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Yellow" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Bob" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "16" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Pink" {
		t.Error(err)
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "Ashley" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "47" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Red" {
		t.Error(err)
		return
	}
}

func TestReadWithComments(t *testing.T) {
	file, err := os.Open("testdata/with-comments.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 data fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "fName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "lName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "dob" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", field.Value, "instead")
	}
	if line.Comment() != "these are headers" {
		t.Error("expect comment in header to be [these are headers] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 fields but got", line.FieldCount(), "instead")
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "john" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "doe" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2023-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", field.Value, "instead")
	}
	if line.Comment() != "this data is probably not accurate" {
		t.Error("expect row", r.CurrentRow(), "to have a comment be [this data is probably not accurate] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 fields but got", line.FieldCount(), "instead")
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "smith" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2024-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "F" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [F] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected to have 0 data fields but got", line.FieldCount(), "instead")
		return
	}
	if line.Comment() != "this is a comment" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", line.Comment(), "instead")
	}

}

func TestReadWithCEmptyLinesAndComments(t *testing.T) {
	file, err := os.Open("testdata/with-empty-lines-and-comments.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "fName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "lName" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "dob" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", field.Value, "instead")
	}
	if line.Comment() != "these are headers" {
		t.Error("expect header 5 to be [these are headers] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 data fields but got", line.FieldCount(), "instead")
		return
	}
	if line.Comment() != "here we go!" {
		t.Error("expect row", r.CurrentRow(), "field ` to be [here we go!] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "john" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "doe" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2023-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", field.Value, "instead")
	}
	if line.Comment() != "this data is probably not accurate" {
		t.Error("expect row", r.CurrentRow(), "field 5 to be [this data is probably not accurate] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected to have 4 fields but got", line.FieldCount(), "instead")
		return
	}
	if field, err := line.Field(0); err != nil || field.Value != "jane" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "smith" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "2024-01-01" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "F" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [F] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected to have 0 data fields but got", line.FieldCount(), "instead")
		return
	}
	if line.Comment() != "this is a comment" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [this is a comment] but got", line.Comment(), "instead")
	}

}

func TestReadInvalidCommentPlacementDueToOmittedFields(t *testing.T) {
	file, err := os.Open("testdata/invalid-comment-placement-due-to-omitted-fields.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "name" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "jersey" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "team" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "sport" {
		t.Error("expect header 4 to be [sport] but got", field.Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "john" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "15" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "tigers" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "baseball" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [baseball] but got", field.Value, "instead")
	}
	line, err = r.Read()
	if err == nil {
		t.Error("expected to return an error but did not", line)
		return
	}

}

func TestReadComplexValues(t *testing.T) {
	file, err := os.Open("testdata/complex-values.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 data fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Country" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.Value != "Capital" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.Value != "Emoji of Flag" {
		t.Error(err)
		return
	}
	if field, err := line.Field(3); err != nil || field.Value != "Interesting Facts" {
		t.Error("expect header 4 to be [Interesting Facts] but got", field.Value, "instead")
	}
	if line.Comment() != "facts generated from Google's Gemini 2024-04-24" {
		t.Error("expected header 5 to be comment and have the value [facts generated from Google's Gemini 2024-04-24] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "France" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Paris" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇫🇷" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "The Eiffel Tower was built for the 1889 World's Fair."+string('\n')+"It was almost torn down afterwards." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The Eiffel Tower was built for the 1889 World's Fair.\\nIt was almost torn down afterwards.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Germany" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Berlin" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇩🇪" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "Germany has over 2,000 beer breweries." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Germany has over 2,000 beer breweries.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Italy" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Rome" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇮🇹" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "The Colosseum in Rome could hold an estimated 50,000 spectators." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The Colosseum in Rome could hold an estimated 50,000 spectators.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 data fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Japan" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Tokyo" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇯🇵🇯🇵" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "Japan is a volcanic archipelago with over 100 active volcanoes."+string('\n')+"The currency is the yen and the symbol is ¥." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Japan is a volcanic archipelago with over 100 active volcanoes.\\nThe currency is the yen and the symbol is ¥.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	if line.Comment() != "has half-width characters" {
		t.Error("expect row", r.CurrentRow(), "field 5 to be a comment but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Spain" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Madrid" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇪🇸" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "Spain has the second highest number of UNESCO World Heritage Sites in the world." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Spain has the second highest number of UNESCO World Heritage Sites in the world.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "United Kingdom" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "London" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇬🇧" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "The United Kingdom is a parliamentary monarchy with a rich history dating back centuries." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The United Kingdom is a parliamentary monarchy with a rich history dating back centuries.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 data fields but got", line.FieldCount(), "instead")
		return
	}

	if line.Comment() != " emphasis on 50 with double quotes" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [ emphasis on 50 with double quotes] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead", line)
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "United States of America" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Washington D.C." {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇺🇸 🏴‍☠️" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || field.Value != "The United States of America is a federal republic with \"50\" states." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The United States of America is a federal republic with \"50\" states.] but got", field.FieldName, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 data fields but got", line.FieldCount(), "instead")
		return
	}

	if line.Comment() != " update the remaining" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [ update the remaining] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "India" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇮🇳" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Errorf("expect row %d field 4 to have value [NULL] but got %+v instead", r.CurrentRow(), field)
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 data fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Canada" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Ottawa" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	if line.Comment() != "need to add facts for the remaining" {
		t.Error("expect row", r.CurrentRow(), "field 5 to have value [need to add facts for the remaining] but got", line.Comment(), "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Australia" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Canberra" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇦🇺" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Brazil" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Brasília" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇧🇷" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Argentina" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Buenos Aires" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇦🇷" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Mexico" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Mexico City" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇲🇽" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "China" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Beijing" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇨🇳" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", line.FieldCount(), "instead")
		return
	}

	if field, err := line.Field(0); err != nil || field.Value != "Russia" {
		t.Error(err)
		return
	}
	if field, err := line.Field(0); err != nil || field.FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(1); err != nil || field.Value != "Moscow" {
		t.Error(err)
		return
	}
	if field, err := line.Field(1); err != nil || field.FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(2); err != nil || field.Value != "🇷🇺" {
		t.Error(err)
		return
	}
	if field, err := line.Field(2); err != nil || field.FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", field.FieldName, "instead")
	}

	if field, err := line.Field(3); err != nil || !field.IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", field.Value, "instead")
	}
	if field, err := line.Field(3); err != nil || field.FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", field.FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", line.FieldCount(), "instead")
		return
	}
	_, err = r.Read()
	if err != io.EOF {
		t.Error("expected an error EOF")
	}

	_, err = r.Read()
	if err != ErrReaderEnded {
		t.Error("expected an error ErrReaderEnded")
	}
}

func TestParseLineTrailingWhiteSpace(t *testing.T) {
	line := `Mexico						"Mexico City"		🇲🇽			  -	`
	fields, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Errorf("expect to have 4 fields but got %+v instead", fields)
		return
	}

	if fields[0].Value != "Mexico" {
		t.Error("field 1 to be [Mexico] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Mexico City" {
		t.Error("field 2 to be [Mexico City] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "🇲🇽" {
		t.Errorf("expect row %d field 3 to have value [NULL] but got %+v instead", 1, fields[2])
	}
	if !fields[3].IsNull {
		t.Error("field 4 to have value [NULL] but got", fields[3].Value, "instead")
	}
}

func TestParseLineWithEmojisAndEscapedDoubleQuotesSurroundedByWhitespace(t *testing.T) {
	line := `"United States of America"  "Washington D.C." 	"🇺🇸 🏴‍☠️"           "The United States of America is a federal republic with ""50"" states."`
	fields, err := parseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line to have 4 fields but got", len(fields), "instead", fields)
		return
	}

	if fields[0].Value != "United States of America" {
		t.Error("field 1 to be [United States of America] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Washington D.C." {
		t.Error("field 2 to be [Washington D.C.] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "🇺🇸 🏴‍☠️" {
		t.Error("field 3 to be [🇺🇸 🏴‍☠️] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "The United States of America is a federal republic with \"50\" states." {
		t.Error("field 4 to have value [The United States of America is a federal republic with \"50\" states.] but got", fields[3].Value, "instead")
	}
}

func TestReaderToDocument(t *testing.T) {
	file, err := os.Open("testdata/complex-values.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	doc, err := r.ToDocument()
	if err != nil {
		t.Error(err)
		return
	}
	if doc.LineCount() != 30 {
		t.Error("Line count is", doc.LineCount())
		return
	}
	line, err := doc.Line(29)
	if err != nil {
		t.Error(err)
		return
	}
	v, err := line.NextField()
	if err != nil {
		t.Error(err)
		return
	}
	if v.Value != "Russia" {
		t.Error("expected the value to be Russia but got", v.Value, "instead")
	}
	line, err = doc.AddLine()
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("South Korea")
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("Seoul")
	if err != nil {
		t.Error(err)
		return
	}

	err = line.AppendNull()
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("Would you've guessed that vodka or gin tops the list? For years, Jinro Soju has been the world's best-selling alcohol! It might not be surprising, given that with 11.2 shots on average, Koreans are also the world's biggest consumer of hard liquor. Haven't been able to try it yet? Time to visit Korea!")
	if err != nil {
		t.Error(err)
		return
	}
	line.UpdateComment("added via document writer")

	data, err := doc.WriteAll()
	if err != nil {
		t.Error(err)
		return
	}
	file, err = os.Create("testdata/complex-output.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		t.Error(err)
	}

}

func TestWriteComplexLine(t *testing.T) {
	strs := []string{}
	strs = append(strs, `Country						Capital    	        "Emoji of Flag" "Interesting Facts" 																								#facts generated from Google's Gemini 2024-04-24`)
	strs = append(strs, `Japan						Tokyo				🇯🇵🇯🇵			"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."   #has half-width characters`)

	r := NewReader(strings.NewReader(strings.Join(strs, "\n")))

	lines, err := r.ReadAll()
	if err != nil {
		t.Error(err)
		return
	}
	if len(lines) != 2 {
		t.Error("expected 2 lines but got", len(lines))
		return
	}
	doc := doc.NewDocument()
	doc.Tabular = true
	doc.SetPadding([]rune{' ', ' '})
	for n, line := range lines {

		ln, err := doc.AddLine()
		if err != nil {
			t.Error(err)
			return
		}
		for fi := range line.FieldCount() {
			field, _ := line.Field(fi)
			if n == 1 && fi == 3 && field.SerializeText() != `"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."` {
				t.Error(field.SerializeText())
				return
			}
			err := ln.Append(field.Value)
			if err != nil {
				t.Error(err)
				return
			}
		}
		if line.Comment() != "" {
			ln.UpdateComment(ln.Comment())
		}
	}
	if doc.LineCount() != 2 {
		t.Error("expected doc to have 2 lines but got", doc.LineCount())
	}

	data, err := doc.WriteAll()
	if err != nil {
		t.Error(err)
		return
	}
	tmpFile := fmt.Sprintf("%ssimple-output.wsv", os.TempDir())
	file, err := os.Create(tmpFile)
	if err != nil {

		t.Error(err)
		return
	}
	t.Log(tmpFile)
	_, err = file.Write(data)
	if err != nil {
		t.Error(err)
	}
}

func TestReadingAndSortingNumberFields(t *testing.T) {
	file, err := os.Open("testdata/unsorted-numbers.wsv")
	if err != nil {
		t.Error(err)
		return
	}
	r := NewReader(file)
	d, err := r.ToDocument()
	if err != nil {
		t.Error(err)
		return
	}
	d.SortBy(doc.SortNumber("Line"))

	expected := [][]string{
		{"Line", "Color", "Last Voted as Favorite"},
		{"1", "Red", "2022-10-29"},
		{"2", "Green", "2023-10-01"},
		{"3", "Purple", "2023-09-18"},
		{"4", "Yellow", "2001-02-28"},
		{"5", "Teal", "2012-01-22"},
		{"6", "Blue", "2024-07-02"},
		{"7", "Brown", "1992-08-05"},
		{"8", "Orange", "2009-06-03"},
		{"9", "Pink", "2003-04-08"},
		{"10", "Fusia", "1963-09-08"},
	}

	for i, line := range d.Lines() {
		for j, field := range line.Fields() {
			expectedValue := expected[i][j]
			if expectedValue != field.Value {
				t.Error("expected", expectedValue, "but got", field.Value)
			}
		}
	}
}
