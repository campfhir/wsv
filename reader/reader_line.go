package reader

import (
	"errors"

	"github.com/campfhir/wsv/record"
)

var (
	ErrFieldNotFound = errors.New("field does not exist")
	ErrEndOfLine     = errors.New("no more fields left in this line")
)

type ReaderLine interface {
	Field(fi int) (*record.RecordField, error)
	// Get the value of comment for the line
	Comment() string
	// Get the line number
	LineNumber() int
	// A count of the number of data fields in the line
	FieldCount() int
	// Get the next field value, or error if at the end of the line for data
	NextField() (*record.RecordField, error)
	// Returns true if the line is a slice of headers
	IsHeaderLine() bool
}

type readerLine struct {
	fields  []record.RecordField
	comment string
	// Lines are 1-indexed
	line int
	// count of data fields, has a getter readerLine.FieldCount()
	fieldCount   int
	currentField int
	isHeaderLine bool
}

func (line *readerLine) NextField() (*record.RecordField, error) {
	if len(line.fields)-1 < line.currentField {
		return nil, ErrEndOfLine
	}
	fieldInd := line.currentField
	line.currentField++
	return &line.fields[fieldInd], nil
}

// Returns the number of data fields, non-comment fields
func (line *readerLine) FieldCount() int {
	return line.fieldCount
}

func (line *readerLine) LineNumber() int {
	return line.line
}

func (line *readerLine) Field(fieldIndex int) (*record.RecordField, error) {
	if len(line.fields)-1 < fieldIndex {
		return nil, ErrFieldNotFound
	}
	return &line.fields[fieldIndex], nil
}

func (line *readerLine) Comment() string {
	return line.comment
}

func (line *readerLine) IsHeaderLine() bool {
	return line.isHeaderLine
}

func (line *readerLine) UpdateComment(val string) {
	line.comment = val
}
