package reader

import (
	"errors"

	"github.com/campfhir/wsv/internal"
)

var (
	ErrFieldNotFound = errors.New("field does not exist")
	ErrEndOfLine     = errors.New("no more fields left in this line")
)

type Line interface {
	// Returns the field value at the 0-index or `ErrFieldNotFound` if out of bounds
	Field(fi int) (*internal.Field, error)
	// Get the value of comment for the line
	Comment() string
	// Get the line number
	LineNumber() int
	// A count of the number of data fields in the line
	FieldCount() int
	// Get the next field value, or error if at the end of the line for data
	NextField() (*internal.Field, error)
	// Returns true if the line is a slice of headers
	IsHeaderLine() bool
	// Returns serialized values of all fields on a line
	FieldsValues() []string
	// All the fields in the line
	Fields() []internal.Field
}

type readerLine struct {
	fields  []internal.Field
	comment string
	// Lines are 1-indexed
	line int
	// count of data fields, has a getter readerLine.FieldCount()
	fieldCount   int
	currentField int
	isHeaderLine bool
}

// A slice of all the fields in this line
func (line *readerLine) Fields() []internal.Field {
	return line.fields
}

// Iterate to the next field in the line
// increments the current field point with each call
//
// If called on an empty line or after all fields have been read will return `ErrEndOfLine`
func (line *readerLine) NextField() (*internal.Field, error) {
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

// The line number for this current line
func (line *readerLine) LineNumber() int {
	return line.line
}

func (line *readerLine) Field(fieldIndex int) (*internal.Field, error) {
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

func (line *readerLine) FieldsValues() []string {
	return internal.Map(line.fields,
		func(e internal.Field, i int, a []internal.Field) string {
			return e.SerializeText()
		})
}
