// Parse a whitespace separated list of values
package reader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"unicode/utf8"

	doc "github.com/campfhir/wsv/document"
	"github.com/campfhir/wsv/internal"
	"github.com/campfhir/wsv/record"
	"github.com/campfhir/wsv/utils"
)

var (
	ErrFieldCount       = errors.New("wrong number of fields")
	ErrLineFeedTerm     = errors.New("line feed terminated before the line end end")
	ErrInvalidNull      = errors.New("null `-` specifier cannot be included without white space surrounding, unless it is the last value in the line. To record a literal `-` please wrap the value in double quotes")
	ErrBareQuote        = errors.New("bare \" in non-quoted-field")
	ErrReaderEnded      = errors.New("reader ended, nothing left to read")
	ErrCommentPlacement = errors.New("comments should be the last elements in a row, if immediate preceding lines are null, they cannot be omitted and must be explicitly declared")
)

type InvalidFieldCountError struct {
	Line    int
	RawLine []byte
	Headers []string
	Fields  []string
}

func (e *InvalidFieldCountError) Error() string {
	v := "values"
	i := "are"

	if len(e.Fields) > len(e.Headers) {
		extra := len(e.Fields) - (len(e.Fields) - len(e.Headers))
		xtra := e.Fields[extra:]
		if len(xtra) == 1 {
			v = "value"
			i = "is"
		}
		return fmt.Sprintf("record on line %d: does not have the correct field count, has %d/%d, the %s (%s) %s extra", e.Line, len(e.Fields), len(e.Headers), v, strings.Join(xtra, ", "), i)
	}
	missing := len(e.Headers) - (len(e.Headers) - len(e.Fields))
	headers := e.Headers[missing:]
	h := "headers"
	if len(headers) == 1 {
		v = "value"
		i = "is"
		h = "header"
	}
	return fmt.Sprintf("record on line %d: does not have the correct field count, has %d/%d, the %s for the %s (%s) %s missing", e.Line, len(e.Fields), len(e.Headers), v, h, strings.Join(record.SerializeValues(headers), "\t"), i)
}

// A ParseError is returned for parsing errors.
// Line numbers are 1-indexed and columns are 0-indexed.
type ParseError struct {
	Line           int   // Line where the error occurred
	FieldPosition  int   // Column (1-based byte index) where the error occurred
	Err            error // The actual error
	Field          *record.Field
	ColumnPosition int
	RawLine        []byte
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s\n%s\nparse error on line %d, column %d: %v", e.RawLine, stringPadLeft("^", e.ColumnPosition), e.Line, e.FieldPosition, e.Err)
}

type ParseErrorCollection struct {
	Errs []error
}

func (e *ParseErrorCollection) Error() string {
	strs := internal.Map(e.Errs, func(x error, i int, a []error) string {
		return x.Error()
	})
	return strings.Join(strs, "\n")
}

// A WSV Document Reader
type Reader struct {
	numLine             int
	offset              int64
	rawBuffer           []byte
	lines               []Line
	headers             []string
	IncludesHeader      bool
	IsTabular           bool
	br                  *bufio.Reader
	NullTrailingColumns bool
	ended               bool
	firstDataRow        int
	AllowPartialError   bool
}

// Returns a slice of headers for a WSV
func (r *Reader) Headers() []string {
	return r.headers
}

// Returns the column name for a given field index, if the index does not exists
// an empty string is returned
func columnName(headers []string, index int) string {
	v, err := utils.GetIndexOfSlice(headers, index)
	if err != nil {
		return ""
	}
	return strings.Clone(*v)
}

func stringPadLeft(str string, length int) string {
	for {
		str = " " + str
		if utf8.RuneCountInString(str) >= length {
			return str[0:length]
		}
	}
}

// Creates a new WSV NewReader
//
// - By default the first non-empty and non-comment line is considered the header
//
// - By default it expects a tabular [each record has the same number of fields] document
//
// - By default omitted trailing fields for a record are allowed
func NewReader(r io.Reader) *Reader {
	return &Reader{
		br:                  bufio.NewReader(r),
		IsTabular:           true,
		IncludesHeader:      true,
		NullTrailingColumns: false,
		lines:               make([]Line, 0),
		ended:               false,
	}
}

// Return the column name at the index i, will return "" if not found
func (r *Reader) ColumnNameOf(i int) (*string, error) {
	return utils.GetIndexOfSlice(r.headers, i)
}

// Return the index of a column name
func (r *Reader) IndexedAt(n string) []int {
	idxs := make([]int, 0)
	for i, h := range r.headers {
		if h != n {
			continue
		}
		idxs = append(idxs, i)
	}
	return idxs

}

// Takes a file path and attempts to read the document as a WSV document
//
// - Will attempt to parse using the default `NewReader()` and return slice of lines it was able to reader
//
// - Can return a *PathError or *ParseError
func Parse(wsvFile string) ([]Line, error) {
	file, err := os.Open(wsvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := NewReader(file)
	records, err := r.ReadAll()
	return records, err
}

// Will read all lines of a reader until it reaches the end of a file or *ParseError
//
// If `err == nil`, it has read the entire document successfully
func (r *Reader) ReadAll() (records []Line, err error) {
	errs := make([]error, 0)
	for {
		record, err := r.Read()
		if err == io.EOF {
			if len(errs) > 0 {
				return records, &ParseErrorCollection{Errs: errs}
			}
			return records, nil
		}
		if err != nil {
			errs = append(errs, err)
		}
		records = append(records, record)
	}
}

type lineField struct {
	Value     string
	IsComment bool
	IsNull    bool
	Col       int
	RawLine   []byte
}

func parseLine(n int, line []byte) ([]lineField, error) {
	var b1 *byte = nil
	var b2 *byte = nil
	var b3 *byte = nil
	var b4 *byte = nil

	doubleQuoted := false

	isNull := false
	startDoubleQuote := 0
	escapedDoubleQuote := 0
	data := []byte{}
	str := make([]lineField, 0)
	// trim the trailing white space from the line
	// line = bytes.TrimRightFunc(line, isFieldDelimiter)
lineLoop:
	for i, b0 := range line {
		if b4 != nil {
			b4 = b3
			b3 = b2
			b2 = b1
			b1 = &b0
		}
		if b4 == nil && b3 != nil {
			b4 = b3
			b3 = b2
			b2 = b1
			b1 = &b0
		}
		if b3 == nil && b2 != nil {
			b3 = b2
			b2 = b1
			b1 = &b0
		}
		if b2 == nil && b1 != nil {
			b2 = b1
			b1 = &b0
		}
		if b1 == nil {
			b1 = &b0
		}
		r := rune(b0)

		switch r {
		case '\n':
			break lineLoop
		case '#':
			if !doubleQuoted {
				if len(line[i:]) < 2 {
					break lineLoop
				}
				data = append(data, line[i+1:]...)
				// since we are copying to the end of line we should remove the suffix of the line feed
				data = bytes.TrimSuffix(data, []byte{'\n'})
				str = append(str, lineField{IsComment: true, Value: string(data), IsNull: isNull, Col: i, RawLine: line})
				// s = ""
				data = []byte{}
				break lineLoop
			}
			data = append(data, byte(r))
			continue
		case '"':
			if bytesToString(b3, b2, b1) == `"/"` {
				data = append(bytes.TrimSuffix(data, []byte{'/'}), byte('\n'))
				continue
			}

			if (b2 == nil || utils.IsFieldDelimiter(rune(*b2))) && !doubleQuoted {
				doubleQuoted = true
				startDoubleQuote = i
				continue
			}

			if (b3 == nil || utils.IsFieldDelimiter(rune(*b3))) && b2 != nil && rune(*b2) == '"' && (len(line)-1 == i || (len(line)-1 > i && utils.IsFieldDelimiter(nextRune(line[i+1:])))) {
				data = []byte{}
				str = append(str, lineField{IsComment: false, Value: string(data), IsNull: isNull, Col: i, RawLine: line})
				doubleQuoted = false
				continue
			}

			if b2 != nil && rune(*b2) == '"' && (b3 == nil || rune(*b3) != '"') && !(len(line)-1 > i+1 && utils.IsFieldDelimiter(nextRune(line[i+1:])) && b3 != nil && rune(*b3) == '/') && !(len(line)-1 > i+2 && nextRune(line[i+1:]) == '/' && nextRune(line[i+2:]) == '"') {
				data = append(data, byte('"'))
				escapedDoubleQuote = i
				continue
			}

			if doubleQuoted && (len(line)-1 == i || (len(line)-1 > i && utils.IsFieldDelimiter(nextRune(line[i+1:])))) && (b2 == nil || rune(*b2) != '"' || i > escapedDoubleQuote) {
				doubleQuoted = false

			}

		case '-':
			if r == '-' && (b2 == nil || utils.IsFieldDelimiter(rune(*b2))) && !doubleQuoted {
				isNull = true
			}
			fallthrough
		default:

			if bytesToString(b3, b2, b1) == `"/"` {
				data = append(bytes.TrimSuffix(data, []byte{'/'}), byte('\n'))
			}
			if isNull && (len(line)-1 == i) {
				str = append(str, lineField{IsComment: false, Value: "", IsNull: isNull, Col: i, RawLine: line})
				break lineLoop
			}
			// currently flagged as null but has more characters left to parse and
			if isNull && len(line)-1 > i && bytes.IndexFunc(line[i:], utils.IsFieldDelimiter) != 1 {
				// the next immediate character is a white space
				if b2 != nil && rune(*b2) == '-' && bytes.IndexFunc([]byte{*b1}, utils.IsFieldDelimiter) == 0 {
					data = []byte{}
				} else {
					// and is not surround by double quotes we have an invalid
					return str, &ParseError{FieldPosition: i, Err: ErrInvalidNull, ColumnPosition: i, Line: n, RawLine: line}
				}

			}

			isDelim := utils.IsFieldDelimiter(r)
			if isDelim && (!doubleQuoted) {
				if len(data) == 0 && !isNull {
					continue
				}
				if string(data) == `"` {
					return str, &ParseError{Line: n, Err: ErrBareQuote, FieldPosition: i, RawLine: line}
				}
				str = append(str, lineField{IsComment: false, Value: string(data), IsNull: isNull, Col: i, RawLine: line})
				isNull = false
				data = []byte{}
				continue
			}
			if isNull && r == '-' {
				// since we identified the field as null and
				continue
			}
			data = append(data, byte(r))
			continue
		}
	}
	if doubleQuoted {
		// the following string value could not be parsed correctly
		return str, &ParseError{FieldPosition: startDoubleQuote, Err: ErrBareQuote, Line: n, RawLine: line, ColumnPosition: startDoubleQuote}
	}
	if len(data) > 0 {
		if string(data) == `"` {
			return str, &ParseError{Line: n, Err: ErrBareQuote, FieldPosition: startDoubleQuote, RawLine: line, ColumnPosition: startDoubleQuote}
		}
		str = append(str, lineField{IsComment: false, Value: string(data), IsNull: isNull, RawLine: line})

	}
	return str, nil
}

func bytesToString(s ...*byte) string {
	str := ""
	for _, b := range s {
		if b == nil {
			continue
		}
		str = str + string(*b)
	}
	return str
}

func (r *Reader) CurrentRow() int {
	return r.numLine
}

// Reads the current line of a reader and returns a *Line.
//
// - If `r.IsTabular == true` and the current line being parsed has more fields than amount of headers
// `r.Read()` returns a the *Line along with the error *ParseError.
//
// - If the record contains a field that cannot be parsed,
// `r.Read()` returns a *Line with as many records as it could before encountering an error.
// The partial record contains all fields read before the error.
//
// - If there is no data left to be read, `r.Read()` returns a *Line with an empty slice Fields and io.EOF.
//
// - Subsequent calls to `r.Read()` after io.EOF returns a nil and ErrReaderEnded
func (r *Reader) Read() (Line, error) {
	var data []byte
	var errRead error
	if r.ended {
		return nil, ErrReaderEnded
	}
	line := readerLine{
		fields:     make([]record.Field, 0),
		fieldCount: 0,
	}
	data, errRead = r.readLine()
	if errRead == io.EOF {
		r.ended = true
		return &line, io.EOF
	}
	line.line = r.numLine

	fields, errRead := parseLine(r.numLine, data)
	if errRead != nil {
		return &line, errRead
	}
	if len(fields) > 0 && r.firstDataRow == 0 && !fields[0].IsComment {
		r.firstDataRow = r.numLine
		if r.IncludesHeader {
			line.isHeaderLine = true
		}
	}

	for i, field := range fields {
		if r.numLine == r.firstDataRow && r.IncludesHeader && !field.IsComment {
			r.headers = append(r.headers, field.Value)
			d := record.Field{Value: field.Value}
			if field.IsNull {
				d.IsNull = true
			}
			d.IsHeader = true
			d.FieldIndex = i
			d.RowIndex = r.numLine
			line.fields = append(line.fields, d)
			line.fieldCount++
			continue
		}
		if field.IsComment {

			// comments must be the first and only value or the last value parsed, if preceding fields are not explicitly defined return an error
			// the exception being non-tabular documents
			if i < len(r.headers) && i != 0 && r.IsTabular {
				return &line, &ParseError{Line: r.numLine, FieldPosition: i + 1, Err: ErrCommentPlacement, ColumnPosition: field.Col, RawLine: field.RawLine}
			}
			line.comment = field.Value
			continue
		}
		line.fieldCount++

		if r.IsTabular && r.IncludesHeader && len(r.headers) < line.fieldCount {
			extraFields := internal.Map(fields[i:], func(e lineField, i int, _ []lineField) string {
				if e.IsNull {
					return "-"
				}
				return record.SerializeValue(e.Value)
			})
			return &line, &InvalidFieldCountError{
				Fields:  append(line.FieldsValues(), extraFields...),
				Headers: r.headers,
				Line:    line.line,
				RawLine: data,
			}
		}

		fieldName := columnName(r.headers, i)
		d := record.Field{Value: field.Value, FieldName: fieldName, IsHeader: false, RowIndex: r.numLine, FieldIndex: i, IsNull: false}
		if field.IsNull {
			d.IsNull = true
			d.Value = ""
		}
		line.fields = append(line.fields, d)
	}

	if len(line.fields) == 0 {
		return &line, errRead
	}

	if len(line.fields) == 0 && len(line.comment) == 0 {
		return &line, errRead
	}

	if r.numLine != 1 && r.NullTrailingColumns && len(line.fields) < len(r.headers) {
		x := len(r.headers) - len(line.fields)
		o := len(line.fields)
		for i := range x {
			h := o + i
			cname := columnName(r.headers, h)
			rec := record.Field{IsNull: true, Value: "", FieldIndex: h, RowIndex: r.numLine, FieldName: cname, IsHeader: false}
			line.fields = append(line.fields, rec)
			line.fieldCount++
		}
	}
	if r.IsTabular && r.IncludesHeader && !r.NullTrailingColumns && len(r.headers) > line.fieldCount {
		return &line, &InvalidFieldCountError{
			Fields:  line.FieldsValues(),
			Headers: r.headers,
			Line:    line.line,
			RawLine: data,
		}
	}
	r.lines = append(r.lines, &line)
	return &line, errRead

}

func nextRune(b []byte) rune {
	r, _ := utf8.DecodeRune(b)
	return r
}

// Reads the current line into a slice bytes
func (r *Reader) readLine() ([]byte, error) {
	line, err := r.br.ReadSlice(utils.CharLineFeed)
	if err == bufio.ErrBufferFull {
		r.rawBuffer = append(r.rawBuffer[:0], line...)
		for err == bufio.ErrBufferFull {
			line, err = r.br.ReadSlice(utils.CharLineFeed)
			r.rawBuffer = append(r.rawBuffer, line...)
		}
		line = r.rawBuffer
	}
	readSize := len(line)
	if readSize > 0 && err == io.EOF {
		err = nil
		// For backwards compatibility, drop trailing \r before EOF.
		if line[readSize-1] == utils.CharCarriageReturn {
			line = line[:readSize-1]
		}
	}
	r.numLine++
	r.offset += int64(readSize)
	if n := len(line); n >= 2 && line[n-2] == utils.CharCarriageReturn && line[n-1] == utils.CharLineFeed {
		line[n-2] = utils.CharLineFeed
		line = line[:n-1]
	}
	// trim the trailing new line
	line = bytes.TrimSuffix(line, []byte("\n"))
	return line, err
}

// Takes a reader an turns that into a document
func (r *Reader) ToDocument() (*doc.Document, error) {
	doc := doc.NewDocument()
	var err error
	var rl Line
	for {
		rl, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return doc, err
		}
		line, err := doc.AddLine()
		if err != nil {
			return nil, err
		}
		if rl.Comment() != "" {
			line.UpdateComment(rl.Comment())
		}
		for i := range rl.FieldCount() {
			field, _ := rl.Field(i)
			if field.IsNull {
				line.AppendNull()
			} else {
				line.Append(field.Value)
			}
		}
	}

	return doc, nil
}
