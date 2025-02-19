// Create a document in the whitespace separated format.
package document

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/campfhir/wsv/record"
	"github.com/campfhir/wsv/utils"
)

type WriteError struct {
	line        int
	fieldIndex  int
	headerCount int

	err error
}

var (
	ErrFieldIndexedNotFound         = errors.New("field does not exist")
	ErrStartedToWrite               = errors.New("document started to write, need to reset document to edit")
	ErrInvalidPaddingRune           = errors.New("only whitespace characters can be used for padding")
	ErrOmitHeaders                  = errors.New("document configured to omit headers")
	ErrLineNotFound                 = errors.New("line does not exist")
	ErrFieldCount                   = errors.New("wrong number of fields")
	ErrCannotSortNonTabularDocument = errors.New("the document is non-tabular and cannot be sorted")
	ErrFieldNotFoundForSortBy       = errors.New("the field was not found")
)

func (e *WriteError) Error() string {

	if e.err == ErrStartedToWrite {
		return fmt.Sprintf("the writer already started and has currently written up until line %d, reset the writer to continue editing the document.", e.line)
	}

	if e.err == ErrFieldIndexedNotFound {
		return fmt.Sprintf("field %d, %s for line %d", e.fieldIndex, e.err.Error(), e.line)
	}

	return e.err.Error()
}

type Document struct {
	Tabular          bool
	EmitHeaders      bool
	lines            []DocumentLine
	maxColumnWidth   map[int]int
	padding          []rune
	currentWriteLine int
	currentField     int
	startedWriting   bool
	headers          []string
	headerLine       int
	hasHeaders       bool
}

func (doc *Document) SetPadding(rs []rune) error {
	for _, r := range rs {
		if !utils.IsFieldDelimiter(r) {
			return &WriteError{err: ErrInvalidPaddingRune}
		}
	}
	doc.padding = rs
	return nil
}

type appendLineField struct {
	val    string
	isNull bool
}

func (d *Document) Lines() []DocumentLine {
	return d.lines
}

func Field(val string) appendLineField {
	return appendLineField{val, false}
}

func Null() appendLineField {
	return appendLineField{"", true}
}

// Adds a line to a document and then appends values to the line added
//
// the literally "-" will be interpreded as null,  if you need a literal "-" use the `line.Append(val string)` function
//
// returns the line that was added, can return an error due validation errors
func (doc *Document) AppendValues(vals ...string) (DocumentLine, error) {
	line, err := doc.AddLine()
	if err != nil {
		return nil, err
	}
	for _, val := range vals {
		if val == "-" {
			err := line.AppendNull()
			if err != nil {
				return line, err
			}
		}
		err := line.Append(val)
		if err != nil {
			return line, err
		}
	}
	return line, nil
}

func (doc *Document) AppendLine(fields ...appendLineField) (DocumentLine, error) {
	line, err := doc.AddLine()
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		if field.isNull {
			err = line.AppendNull()
			if err != nil {
				return line, err
			}
			continue
		}
		err = line.Append(field.val)
		if err != nil {
			return line, err
		}
	}
	return line, nil
}

func (doc *Document) AddLine() (DocumentLine, error) {
	if doc.startedWriting {
		return nil, &WriteError{err: ErrStartedToWrite, line: doc.currentWriteLine}
	}
	pln := len(doc.lines)
	line := documentLine{
		doc:    doc,
		fields: make([]record.RecordField, 0),
		line:   pln + 1,
	}

	doc.lines = append(doc.lines, &line)
	doc.currentField = 0
	return &line, nil
}

// evaluates previous and current record fields and should return true if current field is after previous field
type SortFunc = func(prv *record.RecordField, curr *record.RecordField) bool

type SortOption struct {
	// name of the field in the
	FieldName string
	Desc      bool
}

// Sorts the documents lines in place based on the sort options
//
// Will sort until finished or a field specified is not found, in which case a ErrFieldNotFoundForSortBy is returned
func (doc *Document) SortBy(sortOptions ...SortOption) error {
	if !doc.Tabular {
		return ErrCannotSortNonTabularDocument
	}

	for _, sort := range sortOptions {
		slices.SortStableFunc(doc.lines, func(a DocumentLine, b DocumentLine) int {
			return a.Compare(sort.FieldName, b, sort.Desc)
		})

	}
	doc.ReIndexLineNumbers()
	return nil
}

// Returns the document at the ln specified. Lines are 1-index. If the line does not exist there is an
// ErrLineNotFound error
func (doc *Document) Line(ln int) (DocumentLine, error) {
	if len(doc.lines)-1 < ln-1 || ln < 1 {
		return nil, ErrLineNotFound
	}
	line := doc.lines[ln-1]
	return line, nil
}

func (doc *Document) ResetWrite() {
	doc.startedWriting = false
	doc.currentWriteLine = 0
}

func (doc *Document) WriteLine(n int, includeHeader bool) ([]byte, error) {
	buf := make([]byte, 0)
	if n > len(doc.lines) || n < 1 {
		return buf, ErrLineNotFound
	}

	headers, err := doc.Line(doc.headerLine)
	if err != nil && includeHeader {
		return buf, err
	}

	line := doc.lines[n-1]

	headerLine := make([]string, line.FieldCount())
	dataLine := make([]string, line.FieldCount())
	for i, field := range line.Fields() {
		var header = ""
		if headers != nil {
			headerField, err := headers.Field(i)
			if err == nil {
				header = headerField.SerializeText()
			}
		}
		data := field.SerializeText()
		dl := utf8.RuneCountInString(data)
		hl := utf8.RuneCountInString(header)
		if includeHeader && i < line.FieldCount()-1 {
			if dl >= hl {
				for range dl - hl {
					header = header + " "
				}
			} else {
				for range hl - dl {
					data = data + " "
				}
			}
		}
		headerLine[i] = header
		dataLine[i] = data
	}

	if includeHeader {
		return []byte(strings.Join(headerLine, string(doc.padding)) + "\n" + strings.Join(dataLine, string(doc.padding))), nil
	}
	return []byte(strings.Join(dataLine, string(doc.padding))), nil
}

// Write, writes the currently line to a slice of bytes based on the current line in process, calling write will increment the counter after each successful call.
// Once all lines are process will return will return empty slice, EOF
func (doc *Document) Write() ([]byte, error) {
	doc.startedWriting = true
	buf := make([]byte, 0)

	if len(doc.lines)-1 < doc.currentWriteLine {
		return buf, io.EOF
	}

	line := doc.lines[doc.currentWriteLine]
	if doc.HasHeaders() && !doc.EmitHeaders && doc.currentWriteLine == doc.headerLine {
		return buf, ErrOmitHeaders
	}
	// if configured to be tabular, not an empty line, and has too little/many fields compared to headers return an error
	if doc.Tabular && doc.currentWriteLine != 0 && line.FieldCount() != 0 && line.FieldCount() != len(doc.Headers()) {
		return buf, &WriteError{line: line.LineNumber(), headerCount: len(doc.Headers()), fieldIndex: line.FieldCount(), err: ErrFieldCount}
	}

	for i, field := range line.Fields() {
		mw, err := doc.MaxColumnWidth(i)
		if err != nil {
			continue
		}
		v := field.SerializeText()
		p := utf8.RuneCountInString(v)
		if doc.Tabular && (len(line.Fields())-1 != i) {
			for {
				// pad value with single spaces unless it's the last column or line has a comment
				if p < mw {
					v = fmt.Sprintf("%s%s", v, " ")
					p = utf8.RuneCountInString(v)
					continue
				}
				break
			}
		}

		if i == 0 {
			buf = append(buf, []byte(v)...)
		} else {
			buf = append(buf, utils.RuneToBytes(doc.padding)...)
			buf = append(buf, []byte(v)...)
		}
	}
	if len(line.Comment()) > 0 {
		if len(buf) > 0 {
			buf = append(buf, utils.RuneToBytes(doc.padding)...)
			buf = append(buf, []byte(fmt.Sprintf("#%s", line.Comment()))...)
		} else {
			buf = append(buf, []byte(fmt.Sprintf("#%s", line.Comment()))...)

		}
	}
	buf = append(buf, byte('\n'))
	doc.currentWriteLine += 1
	return buf, nil
}

func (doc *Document) WriteAll() ([]byte, error) {
	data := make([]byte, 0)
	for {
		d, err := doc.Write()
		if err == io.EOF {
			break
		}
		if err != nil {

			return data, err
		}
		data = append(data, d...)
	}
	return data, nil
}

func (doc *Document) LineCount() int {
	return len(doc.lines)
}

// Returns a comment if one exists for the rows or an error if comment does not exist
// lines are 1-indexed
func (doc *Document) CommentFor(ln int) (string, error) {
	if len(doc.lines) < ln {
		return "", fmt.Errorf("there are no records found for row %d, please ensure you are indexing as 1-indexed values", ln)
	}
	line := doc.lines[ln-1]

	if len(line.Comment()) > 0 {
		return line.Comment(), nil
	}
	msg := fmt.Errorf("comment not found for row %d", ln)
	return "", msg
}

func (doc *Document) CalculateMaxFieldLengths() {
	for _, line := range doc.lines {
		if line == nil {
			continue
		}
		for fieldInd, field := range line.Fields() {
			fw := field.CalculateFieldLength()
			doc.SetMaxColumnWidth(fieldInd, fw)
		}
	}
}

func (doc *Document) HasHeaders() bool {
	return doc.hasHeaders
}

func (doc *Document) SetMaxColumnWidth(col int, len int) {
	v, ok := doc.maxColumnWidth[col]
	if !ok {
		doc.maxColumnWidth[col] = len
		return
	}
	if v < len {
		doc.maxColumnWidth[col] = len
	}
}

func (doc *Document) MaxColumnWidth(col int) (int, error) {
	v, ok := doc.maxColumnWidth[col]
	if !ok {
		return 0, ErrFieldIndexedNotFound
	}
	return v, nil
}

func (doc *Document) SetHideHeaderStyle(v bool) {
	if !doc.hasHeaders {
		return
	}
	doc.EmitHeaders = v
}

func (doc *Document) Headers() []string {
	return doc.headers
}

func (doc *Document) UpdateHeader(fi int, val string) error {
	if !doc.HasHeaders() {
		return nil
	}

	for i := range doc.LineCount() {
		line, _ := doc.Line(i)
		for fi := range line.FieldCount() {
			if fi == 0 {
				line.UpdateField(fi, val)
			}
			err := line.UpdateFieldName(fi, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (doc *Document) AppendHeader(val string) {
	doc.headers = append(doc.headers, val)
}

func Fields(s ...string) []appendLineField {
	fields := make([]appendLineField, len(s))
	for i, v := range s {
		fields[i] = Field(v)
	}
	return fields
}

func (doc *Document) ReIndexLineNumbers() {
	for i, line := range doc.lines {
		line.ReIndexLineNumber(i + 1)
	}
}

func (doc *Document) WriteAllTo(w io.Writer) error {
	for {
		d, err := doc.Write()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		w.Write(d)
	}
	return nil
}

func NewDocument() *Document {
	doc := Document{
		Tabular:          true,
		EmitHeaders:      true,
		lines:            make([]DocumentLine, 0),
		currentWriteLine: 0,
		currentField:     0,
		maxColumnWidth:   make(map[int]int, 0),
		headerLine:       0,
		startedWriting:   false,
		// The runes in between data values
		padding:    []rune{' ', ' '},
		headers:    make([]string, 0),
		hasHeaders: true,
	}
	return &doc
}
