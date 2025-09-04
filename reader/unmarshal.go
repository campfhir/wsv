package reader

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/campfhir/wsv/internal"
)

type unmarshalError struct {
	field      string
	format     string
	comment    bool
	fieldIndex []int
	fieldType  string
	cause      error
}

type UnmarshalWSV interface {
	UnmarshalWSV(value string, format string) error
}

func (e *unmarshalError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("unmarshal error field: '%s' format: [%s], field index: %d, field type: %s caused by %s", e.field, e.format, e.fieldIndex, e.fieldType, e.cause)
	}
	return fmt.Sprintf("unmarshal error field: '%s' attributes: [%s], field index: %d, field type: %s", e.field, e.format, e.fieldIndex, e.fieldType)
}

func unmarshalRow(fields []internal.Field, t reflect.Type) (*reflect.Value, error) {
	val := reflect.New(t).Elem()
	if val.Kind() != reflect.Struct {
		return nil, errors.New("expected a struct to unmarshal to")
	}

	// Collect all fields indexed by tag name
	tagLookup := make(map[string]fieldInfo)
	collectFields(val.Type(), nil, tagLookup)

	for _, field := range fields {
		fi, ok := tagLookup[field.FieldName]
		if !ok || !fi.Field.IsExported() {
			continue
		}

		key, _, format, literalEmptyField := parseWSVTag(fi.Field)
		if key == "-" && !literalEmptyField {
			continue
		}
		if key == "" || key != field.FieldName {
			continue
		}

		sf := val.FieldByIndex(fi.Index)

		// Custom Unmarshaler
		if sf.CanAddr() {
			if u, ok := sf.Addr().Interface().(UnmarshalWSV); ok {
				if err := u.UnmarshalWSV(field.Value, format); err != nil {
					return nil, newUnmarshalError(key, format, fi.Index, sf.Type().String(), err)
				}
				continue
			}
		}

		if err := setValue(sf, field, key, format, fi.Index); err != nil {
			return nil, err
		}
	}

	return &val, nil
}

// setValue assigns a field value according to its kind/pointer type.
func setValue(sf reflect.Value, field internal.Field, fieldName, format string, idx []int) error {
	switch sf.Kind() {
	case reflect.String:
		sf.SetString(field.Value)
	case reflect.Bool:
		format = defaultIfEmpty(format, "True|False")
		return setBool(sf, field.Value, fieldName, format, idx)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(sf, field.Value, fieldName, format, idx)
	case reflect.Float32, reflect.Float64:
		return setFloat(sf, field.Value, fieldName, format, idx)
	case reflect.Ptr:
		if field.IsNull {
			return nil
		}
		return setPointer(sf, field, fieldName, format, idx)
	default:
		return newUnmarshalError(fieldName, format, idx, sf.Type().String(), nil)
	}
	return nil
}

func setBool(sf reflect.Value, raw, field, format string, idx []int) error {
	v, err := internal.ParseBool(raw, format)
	if err != nil {
		return newUnmarshalError(field, format, idx, sf.Kind().String(), err)
	}
	sf.SetBool(v)
	return nil
}

var intMap = map[string]int64{
	"base2":  int64(2),
	"base8":  int64(8),
	"base10": int64(10),
	"base16": int64(16),
}

func parseInt(raw, format string) (v int64, err error) {
	base := int64(10)
	if b, ok := intMap[format]; ok {
		base = b
		format = ""
	}
	if format != "" {
		base, err = strconv.ParseInt(format, 10, strconv.IntSize)
	}

	if err != nil {
		return
	}
	v, err = strconv.ParseInt(raw, int(base), strconv.IntSize)
	return
}

func setInt(sf reflect.Value, raw, field string, format string, idx []int) error {
	v, err := parseInt(raw, format)
	if err != nil {
		return newUnmarshalError(field, format, idx, sf.Kind().String(), err)
	}
	sf.SetInt(v)
	return nil
}

func setFloat(sf reflect.Value, raw, field string, format string, idx []int) error {
	v, err := strconv.ParseFloat(raw, strconv.IntSize)
	if err != nil {
		return newUnmarshalError(field, format, idx, sf.Kind().String(), err)
	}
	sf.SetFloat(v)
	return nil
}

func setPointer(sf reflect.Value, field internal.Field, fieldName, format string, idx []int) error {
	switch sf.Type().Elem().Kind() {
	case reflect.String:
		sf.Set(reflect.ValueOf(&field.Value))
	case reflect.Bool:
		format = defaultIfEmpty(format, "True|False")
		v, err := internal.ParseBool(field.Value, format)
		if err != nil {
			return newUnmarshalError(fieldName, format, idx, "*bool", err)
		}
		sf.Set(reflect.ValueOf(&v))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := parseInt(field.Value, format)
		if err != nil {
			return newUnmarshalError(fieldName, format, idx, sf.Type().String(), err)
		}
		sf.Set(reflect.New(sf.Type().Elem()))
		sf.Elem().SetInt(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(field.Value, 64)
		if err != nil {
			return newUnmarshalError(fieldName, format, idx, sf.Type().String(), err)
		}
		sf.Set(reflect.New(sf.Type().Elem()))
		sf.Elem().SetFloat(v)
	default:
		// Handle special cases like *time.Time
		if sf.Type() == reflect.TypeOf(&time.Time{}) {
			v, err := time.Parse(format, field.Value)
			if err != nil {
				return newUnmarshalError(fieldName, format, idx, "*time.Time", err)
			}
			sf.Set(reflect.ValueOf(&v))
			return nil
		}
		return newUnmarshalError(fieldName, format, idx, sf.Type().String(), nil)
	}
	return nil
}

func newUnmarshalError(field string, format string, idx []int, typ string, cause error) error {
	return &unmarshalError{
		field:      field,
		format:     format,
		fieldIndex: idx,
		fieldType:  typ,
		cause:      cause,
	}
}

type fieldInfo struct {
	Index []int
	Field reflect.StructField
}

// collectFields flattens all fields (including embedded) with tag info
func collectFields(t reflect.Type, parentIndex []int, tags map[string]fieldInfo) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		idx := append(parentIndex, i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			// recurse into embedded struct
			collectFields(f.Type, idx, tags)
			continue
		}

		tag, _ := f.Tag.Lookup("wsv")
		if tag == "" {
			tag = f.Name
		}
		keys := internal.SplitQuoted(tag)
		if len(keys) == 0 || keys[0] == "" {
			continue
		}

		tags[keys[0]] = fieldInfo{Index: idx, Field: f}
	}
}

// Unmarshal a slice of bytes into a struct `v`.
//
// Will use the struct tag `wsv` to unmarshal the input.
func Unmarshal(d []byte, v any) error {
	r := NewReader(strings.NewReader(string(d)))
	vt := reflect.TypeOf(v)
	sl := reflect.ValueOf(v)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
		sl = sl.Elem()
	}

	switch vt.Kind() {
	case reflect.Slice:
		vt = vt.Elem()
	default:
		return errors.New("can only unmarshal into a slice")
	}

	for {
		rl, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if rl.IsHeaderLine() {
			continue
		}

		fields := rl.Fields()
		val, err := unmarshalRow(fields, vt)
		if err != nil {
			return err
		}
		n := reflect.Append(sl, *val)
		sl.Set(n)
	}

	return nil
}
