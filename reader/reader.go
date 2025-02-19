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

// A ParseError is returned for parsing errors.
// Line numbers are 1-indexed and columns are 0-indexed.
type ParseError struct {
	Line          int   // Line where the error occurred
	Column        int   // Column (1-based byte index) where the error occurred
	Err           error // The actual error
	NeighborBytes []byte
}

func (e *ParseError) Error() string {
	if e.Err == ErrFieldCount {
		return fmt.Sprintf("record on line %d: %v", e.Line, e.Err)
	}
	return fmt.Sprintf("parse error on line %d, column %d [%s]: %v", e.Line, e.Column, string(e.NeighborBytes), e.Err)

}

// These are the errors that can be returned in ParseError.Err.

type Reader struct {
	numLine             int
	offset              int64
	rawBuffer           []byte
	FieldsPerRecord     int
	lines               []ReaderLine
	headers             []string
	IncludesHeader      bool
	IsTabular           bool
	r                   *bufio.Reader
	NullTrailingColumns bool
	ended               bool
	firstDataRow        int
}

func (r *Reader) Headers() []string {
	return r.headers
}

func columnName(headers []string, index int) string {
	v, err := utils.GetIndexOfSlice(headers, index)
	if err != nil {
		return ""
	}
	return strings.Clone(*v)
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:                   bufio.NewReader(r),
		IsTabular:           true,
		IncludesHeader:      true,
		NullTrailingColumns: true,
		lines:               make([]ReaderLine, 0),
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

func Parse(wsvFile string) ([]ReaderLine, error) {
	file, err := os.Open(wsvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := NewReader(file)
	records, err := r.ReadAll()
	return records, err
}

func (r *Reader) ReadAll() (records []ReaderLine, err error) {
	for {
		record, err := r.Read()
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
}

type LineField struct {
	Value     string
	IsComment bool
	IsNull    bool
}

func ParseLine(n int, line []byte) ([]LineField, error) {
	var b1 *byte = nil
	var b2 *byte = nil
	var b3 *byte = nil
	var b4 *byte = nil

	doubleQuoted := false

	isNull := false
	startDoubleQuote := 0
	escapedDoubleQuote := 0
	data := []byte{}
	str := make([]LineField, 0)
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
			if i < len(line)-1 {
				d := neighborBytes(i, line)
				return str, &ParseError{Line: n, Column: i, Err: ErrLineFeedTerm, NeighborBytes: d}
			}
			break lineLoop
		case '#':
			if !doubleQuoted {
				if len(line[i:]) < 2 {
					break lineLoop
				}
				data = append(data, line[i+1:]...)
				// since we are copying to the end of line we should remove the suffix of the line feed
				data = bytes.TrimSuffix(data, []byte{'\n'})
				str = append(str, LineField{IsComment: true, Value: string(data), IsNull: isNull})
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
				str = append(str, LineField{IsComment: false, Value: string(data), IsNull: isNull})
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
				str = append(str, LineField{IsComment: false, Value: "", IsNull: isNull})
				break lineLoop
			}
			// currently flagged as null but has more characters left to parse and
			if isNull && len(line)-1 > i && bytes.IndexFunc(line[i:], utils.IsFieldDelimiter) != 1 {
				// the next immediate character is a white space
				if b2 != nil && rune(*b2) == '-' && bytes.IndexFunc([]byte{*b1}, utils.IsFieldDelimiter) == 0 {
					data = []byte{}
				} else {
					// and is not surround by double quotes we have an invalid
					return str, &ParseError{Column: i, Err: ErrInvalidNull}
				}

			}

			isDelim := utils.IsFieldDelimiter(r)
			if isDelim && (!doubleQuoted) {
				if len(data) == 0 && !isNull {
					continue
				}
				if string(data) == `"` {
					nb := neighborBytes(i, line)
					return str, &ParseError{Line: n, Err: ErrBareQuote, Column: i, NeighborBytes: nb}
				}
				str = append(str, LineField{IsComment: false, Value: string(data), IsNull: isNull})
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

		nb := neighborBytes(startDoubleQuote, line)
		return str, &ParseError{Column: startDoubleQuote, Err: ErrBareQuote, Line: n, NeighborBytes: nb}
	}
	if len(data) > 0 {
		if string(data) == `"` {
			nb := neighborBytes(startDoubleQuote, line)
			return str, &ParseError{Line: n, Err: ErrBareQuote, Column: startDoubleQuote, NeighborBytes: nb}
		}
		str = append(str, LineField{IsComment: false, Value: string(data), IsNull: isNull})

	}
	return str, nil
}

func neighborBytes(i int, line []byte) (neighbor []byte) {
	if i < 0 {
		return neighbor
	}
	p := 5 - (5 - i)
	s := 5 - (5 - i)
	if p > 5 {
		p = 5
	}
	if s > 5 {
		s = 5
	}
	if len(line[i:]) < i+s {
		s = (len(line[i:]) - 1) + i
	}
	neighbor = line[p:s]
	return neighbor
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

// Read a slice of RecordField from r.
// If the reader IsTabular and the row being parsed has more
// fields than the header row will return the records, ParseError
// Read returns the record along with the error ParseError.
// If the record contains a field that cannot be parsed,
// Read returns a partial record along with the parse error.
// The partial record contains all fields read before the error.
// If there is no data left to be read, Read returns an empty RecordField slice, io.EOF.
// Subsequent calls to Read after io.EOF returns an empty RecordFieldSlice, ErrReaderEnded
func (r *Reader) Read() (ReaderLine, error) {
	var data []byte
	var errRead error
	line := readerLine{
		fields:     make([]record.RecordField, 0),
		fieldCount: 0,
	}
	if r.ended {
		return &line, ErrReaderEnded
	}
	data, errRead = r.readLine()
	if errRead == io.EOF {
		r.ended = true
		return &line, io.EOF
	}
	line.line = r.numLine

	fields, errRead := ParseLine(r.numLine, data)
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
			d := record.RecordField{Value: field.Value}
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
			if i < len(r.headers) && i != 0 {
				return &line, &ParseError{Line: r.numLine, Column: 0, Err: ErrCommentPlacement}
			}
			line.comment = field.Value
			continue
		}
		line.fieldCount++

		if r.IsTabular && r.IncludesHeader && len(r.headers) < line.fieldCount {
			return &line, &ParseError{Line: r.numLine, Column: 0, Err: ErrFieldCount}
		}
		fieldName := columnName(r.headers, i)
		d := record.RecordField{Value: field.Value, FieldName: fieldName, IsHeader: false, RowIndex: r.numLine, FieldIndex: i, IsNull: false}
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
			rec := record.RecordField{IsNull: true, Value: "", FieldIndex: h, RowIndex: r.numLine, FieldName: cname, IsHeader: false}
			line.fields = append(line.fields, rec)
			line.fieldCount++
		}
	}
	r.lines = append(r.lines, &line)
	return &line, errRead

}

func nextRune(b []byte) rune {
	r, _ := utf8.DecodeRune(b)
	return r
}

func (r *Reader) readLine() ([]byte, error) {
	line, err := r.r.ReadSlice(utils.CharLineFeed)
	if err == bufio.ErrBufferFull {
		r.rawBuffer = append(r.rawBuffer[:0], line...)
		for err == bufio.ErrBufferFull {
			line, err = r.r.ReadSlice(utils.CharLineFeed)
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

func (r *Reader) ToDocument() (*doc.Document, error) {
	doc := doc.NewDocument()
	var err error
	var rl ReaderLine
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
		for i := range rl.FieldCount() {
			field, _ := rl.Field(i)
			line.Append(field.Value)
		}
	}

	return doc, nil
}
