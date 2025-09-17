package document

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/campfhir/wsv/internal"
)

var (
	ErrUnsupportMarshalType = errors.New("unsupported type to marshal")
	ErrNoDataMarshalled     = errors.New("no data marshalled")
)

type MarshalWSV interface {
	MarshalWSV(format string) (*string, error)
}

type row struct {
	fields  []internal.Field
	comment string
}

func marshalRow(v reflect.Value) (*row, error) {
	fields := make([]internal.Field, 0, v.NumField())
	var comment string

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := v.Type().Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}

		// Parse tag
		key, isComment, format, literalEmptyField := internal.ParseWSVTag(fieldType)
		if key == "-" && !literalEmptyField {
			continue
		}
		// Custom marshaller
		if fieldValue.CanAddr() {
			if u, ok := fieldValue.Addr().Interface().(MarshalWSV); ok {
				val, isNull, err := callCustomMarshaller(u, format)
				if err != nil {
					return nil, err
				}
				if isComment {
					comment = appendComment(comment, val)
					continue
				}
				fields = append(fields, internal.Field{
					FieldName:  key,
					Value:      val,
					IsNull:     isNull,
					FieldIndex: i,
				})
				continue
			}
		}

		// Deref pointers/interfaces
		fieldValue, isNil := deref(fieldValue)
		if isNil {
			if isComment {
				continue
			}
			fields = append(fields, internal.Field{FieldName: key, IsNull: true, FieldIndex: i})
			continue
		}

		// Handle concrete kinds
		switch fieldValue.Kind() {
		case reflect.String:
			val := fieldValue.String()
			if isComment {
				comment = appendComment(comment, val)
				continue
			}
			fields = append(fields, internal.Field{FieldName: key, Value: val, FieldIndex: i})

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if t, ok := fieldValue.Interface().(time.Duration); ok {
				val := t.String()
				if isComment {
					comment = appendComment(comment, val)
					continue
				}
				fields = append(fields, internal.Field{FieldName: key, Value: val, FieldIndex: i})
				continue
			}
			format = internal.DefaultIfEmpty(format, "%d")
			val := fmt.Sprintf(format, fieldValue.Int())
			if isComment {
				comment = appendComment(comment, val)
				continue
			}
			fields = append(fields, internal.Field{FieldName: key, Value: val, FieldIndex: i})

		case reflect.Float32, reflect.Float64:
			format = internal.DefaultIfEmpty(format, "%.2f")
			val := fmt.Sprintf(format, fieldValue.Float())
			if isComment {
				comment = appendComment(comment, val)
				continue
			}
			fields = append(fields, internal.Field{FieldName: key, Value: val, FieldIndex: i})

		case reflect.Bool:
			format = internal.DefaultIfEmpty(format, "True|False")
			val := internal.FormatBool(fieldValue.Bool(), format)
			if isComment {
				comment = appendComment(comment, val)
				continue
			}
			fields = append(fields, internal.Field{FieldName: key, Value: val, FieldIndex: i})

		case reflect.Struct:
			if t, ok := fieldValue.Interface().(time.Time); ok {
				format = internal.ParseStructTagDateFormat(format)
				val := t.Format(format)
				if isComment {
					comment = appendComment(comment, val)
					continue
				}
				fields = append(fields, internal.Field{FieldName: key, Value: val, FieldIndex: i})
				continue
			}
			fallthrough
		default:
			if key != "" {
				return nil, fmt.Errorf("does not support '%s' to marshal without implementing [MarshalWSV]", fieldValue.Type().String())
			}
		}
	}

	return &row{fields, comment}, nil
}

func callCustomMarshaller(u MarshalWSV, format string) (val string, isNull bool, err error) {
	v, err := u.MarshalWSV(format)
	if err != nil {
		return
	}
	if v == nil {
		isNull = true
		return
	}
	val = *v
	return
}

// Dereference pointers/interfaces safely
func deref(v reflect.Value) (reflect.Value, bool) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return v, true
		}
		return v.Elem(), false
	}
	return v, false
}

func appendComment(existing, newVal string) string {
	if existing == "" {
		return newVal
	}
	return existing + " " + newVal
}

// Marshal returns a WSV encoding of s with the option to sort by columns.
//
// Marsal iterates over the elements of s. For each element of s it iterates over the fields of in the s[n].
//
// If a field in the s[n] implements [MarshalWSV] it will call the [MarshalWSV.MarshalWSV()]. Otherwise it will
// use the default implementation.
//
// Marshal uses the `wsv` tag of fields within s[n] in following format: `wsv:"[field name][,format:[[string]][,comment]"`
//
// `field name` attribute can be empty and will take the name of the exported field.
//
// `format:` uses the value immediately after the colon `:`, until the end of the tag or a comma `,` unless enclosed in single quotes `'`.
//
// Struct tags with the `comment` attribute will be appended to the end of the line as comment. Subsequent fields with the `comment` attribute will be appended to previous comment
// on the same line in the order declared in the struct with a single space ` ` between each field. Empty or nil values will not be appended.
// The `field name` is ignored but must contain a comma `,` before.
//
// Example:
//
//	type User struct {
//	  LastLogin time.Time `wsv:",comment"`
//	  Points float64 `wsv:",format:%.4f,comment"`
//	}
//
// Values will parsed with their formats or call [MarshalWSV.MarshalWSV()] prior to appending comments. See below for details about the `format`
//
// All exported fields in the struct s[n] will try to marshal unless a specific `wsv` tag with a field name of `-` is provided.
// If the field name name is expect to literally be `-` there needs to be comma `,` to follow.
//
// Supports `string`, `int`, `bool`, `float`, `time.Time`.
//
// Fields with the type of `string` do not support the `format:` attribute in the struct tag and will just be ignored if specified.
//
// Fields with the type of `int` can alter their byte representation with the `format:` attribute in the struct tag.
// The format is in the format of `fmt.Sprintf` and the default is `%d`.
//
// Example:
//
//	type Person struct {
//	   Age int `wsv:"age,format:%d"`
//	}
//
// Fields with type `bool` can also alter their byte representation with the `format:` attribute in the struct tag.
// The format is template `true|false` with the left side of the `|` representing the literal value of `true` and the right side representing the literal value of `false`. The default is `True|False`.
//
// Example:
//
//	type User struct {
//	  IsAdmin bool `wsv:"Admin,format:yes|no"`
//	}
//
// Field with type `float` can alter their byte representation with the `format:` attribute in the struct tag.
// The format is in the format of `fmt.Sprintf` and the default is `%.2f`
//
//	type Employee struct {
//	  Salary float32 `wsv:"Weekly Salary,format:%.2f"`
//	}
//
// Field with type `time.Time` can alter their byte representation with the `format:` attribute in the struct tag.
// The format is in the format of `time.Format` and the default is `time.RFC3339`
// The time can be written a literal string layout `2006-01-02` or using a the following shorthand values:
//
// * layout
//
// * ansic
//
// * unixdate
//
// * rubydate
//
// * rfc822
//
// * rfc822z
//
// * rfc850
//
// * rfc1123
//
// * rfc1123z
//
// * rfc3339
//
// * rfc3339nano
//
// * kitchen
//
// * stamp
//
// * stampmilli
//
// * stampmicro
//
// * stampnano
//
// * datetime
//
// * dateonly
//
// * date
//
// * timeonly
//
// * time
//
// If you need a format with `,` you can escape the `,` by wrapping the `format` in `'` single qoutes. For example `wsv:"field,format:'Jan 02, 2006'`.
//
// Example:
//
//	type TimeOff struct {
//	  Date *time.Time `wsv:"PTO,format:'January 02, 2006'"`
//	  Requested time.Time `wsv:"Requested,format:2006-01-02"`
//	  Approved *time.Time `wsv:,format:rfc3339"`
//	}
func MarshalWithOptions[T any](s []T, options ...*internal.SortOption) ([]byte, error) {
	v_ := reflect.ValueOf(s)
	t_ := reflect.TypeOf(s)
	var rows []row
	switch t_.Kind() {
	case reflect.Slice:
		for i := range v_.Len() {
			v := v_.Index(i)
			row, err := marshalRow(v)
			if err != nil {
				return nil, err
			}
			rows = append(rows, *row)
		}
	default:
		return nil, ErrUnsupportMarshalType
		// v_ = v_.Len()
	}

	if len(rows) <= 0 {
		return nil, ErrNoDataMarshalled
	}
	doc := NewDocument()
	line, err := doc.AddLine()
	if err != nil {
		return nil, err
	}
	for _, field := range rows[0].fields {
		if err = line.Append(field.FieldName); err != nil {
			return nil, err
		}
	}

	for _, row := range rows {
		line, err = doc.AddLine()
		if err != nil {
			return nil, err
		}
		for _, field := range row.fields {
			if field.IsNull {
				if err = line.AppendNull(); err != nil {
					return nil, err
				}
				continue
			}
			if err = line.Append(field.Value); err != nil {
				return nil, err
			}
		}
		line.UpdateComment(row.comment)
	}
	err = doc.SortBy(options...)
	if err != nil {
		return nil, err
	}
	d, err := doc.WriteAll()
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Marshal returns a WSV encoding of s.
//
// Marsal iterates over the elements of s. For each element of s it iterates over the fields of in the s[n].
//
// If a field in the s[n] implements [MarshalWSV] it will call the [MarshalWSV.MarshalWSV()]. Otherwise it will
// use the default implementation.
//
// Marshal uses the `wsv` tag of fields within s[n] in following format: `wsv:"[field name][,format:[[string]][,comment]"`
//
// `field name` attribute can be empty and will take the name of the exported field.
//
// `format:` uses the value immediately after the colon `:`, until the end of the tag or a comma `,` unless enclosed in single quotes `'`.
//
// Struct tags with the `comment` attribute will be appended to the end of the line as comment. Subsequent fields with the `comment` attribute will be appended to previous comment
// on the same line in the order declared in the struct with a single space ` ` between each field. Empty or nil values will not be appended.
// The `field name` is ignored but must contain a comma `,` before.
//
// Example:
//
//	type User struct {
//	  LastLogin time.Time `wsv:",comment"`
//	  Points float64 `wsv:",format:%.4f,comment"`
//	}
//
// Values will parsed with their formats or call [MarshalWSV.MarshalWSV()] prior to appending comments. See below for details about the `format`
//
// All exported fields in the struct s[n] will try to marshal unless a specific `wsv` tag with a field name of `-` is provided.
// If the field name name is expect to literally be `-` there needs to be comma `,` to follow.
//
// Supports `string`, `int`, `bool`, `float`, `time.Time`.
//
// Fields with the type of `string` do not support the `format:` attribute in the struct tag and will just be ignored if specified.
//
// Fields with the type of `int` can alter their byte representation with the `format:` attribute in the struct tag.
// The format is in the format of `fmt.Sprintf` and the default is `%d`.
//
// Example:
//
//	type Person struct {
//	   Age int `wsv:"age,format:%d"`
//	}
//
// Fields with type `bool` can also alter their byte representation with the `format:` attribute in the struct tag.
// The format is template `true|false` with the left side of the `|` representing the literal value of `true` and the right side representing the literal value of `false`. The default is `True|False`.
//
// Example:
//
//	type User struct {
//	  IsAdmin bool `wsv:"Admin,format:yes|no"`
//	}
//
// Field with type `float` can alter their byte representation with the `format:` attribute in the struct tag.
// The format is in the format of `fmt.Sprintf` and the default is `%.2f`
//
//	type Employee struct {
//	  Salary float32 `wsv:"Weekly Salary,format:%.2f"`
//	}
//
// Field with type `time.Time` can alter their byte representation with the `format:` attribute in the struct tag.
// The format is in the format of `time.Format` and the default is `time.RFC3339`
// The time can be written a literal string layout `2006-01-02` or using a the following shorthand values:
//
// * layout
//
// * ansic
//
// * unixdate
//
// * rubydate
//
// * rfc822
//
// * rfc822z
//
// * rfc850
//
// * rfc1123
//
// * rfc1123z
//
// * rfc3339
//
// * rfc3339nano
//
// * kitchen
//
// * stamp
//
// * stampmilli
//
// * stampmicro
//
// * stampnano
//
// * datetime
//
// * dateonly
//
// * date
//
// * timeonly
//
// * time
//
// If you need a format with `,` you can escape the `,` by wrapping the `format` in `'` single qoutes. For example `wsv:"field,format:'Jan 02, 2006'`.
//
// Example:
//
//	type TimeOff struct {
//	  Date *time.Time `wsv:"PTO,format:'January 02, 2006'"`
//	  Requested time.Time `wsv:"Requested,format:2006-01-02"`
//	  Approved *time.Time `wsv:,format:rfc3339"`
//	}
func Marshal[T any](s []T) ([]byte, error) {
	return MarshalWithOptions(s, nil)
}
