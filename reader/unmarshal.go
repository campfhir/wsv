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
	utils "github.com/campfhir/wsv/internal"
)

type unmarshalError struct {
	tag        string
	attributes []string
	fieldIndex []int
	fieldType  string
	cause      error
}

type UnmarshalWSV interface {
	UnmarshalWSV(value string, format string) error
}

func (e *unmarshalError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("unmarshal error tag: '%s' attributes: [%s], field index: %d, field type: %s caused by %s", e.tag, strings.Join(e.attributes, ","), e.fieldIndex, e.fieldType, e.cause)
	}
	return fmt.Sprintf("unmarshal error tag: '%s' attributes: [%s], field index: %d, field type: %s", e.tag, strings.Join(e.attributes, ","), e.fieldIndex, e.fieldType)
}

func unmarshalRow(fields []utils.Field, t reflect.Type) (*reflect.Value, error) {
	val := reflect.New(t).Elem()
	_type := val.Type()
	if _type.Kind() != reflect.Struct {
		return nil, errors.New("expected a struct to unmarshal to")
	}
	tagLookUp := make(map[string]fieldInfo)
	collectFields(_type, nil, tagLookUp)

fl:
	for fi, field := range fields {
		_ = fi
		fi, ok := tagLookUp[field.FieldName]
		if !ok {
			continue
		}
		ty := fi.Field

		if !ty.IsExported() {
			continue
		}
		tag, _ := ty.Tag.Lookup("wsv")
		if tag == "" {
			tag = ty.Name
		}
		name := ""
		format := ""
		keys := utils.SplitQuoted(tag)
		if len(keys) > 0 {
			name = keys[0]
		}
		if name == "" || name != field.FieldName {
			continue
		}
		for _, key := range keys {
			if f, ok := strings.CutPrefix(key, "format:"); ok {
				format = f
				break
			}
		}

		sf := val.FieldByIndex(fi.Index)
		if sf.CanAddr() {
			addr := sf.Addr().Interface()
			if u, ok := addr.(UnmarshalWSV); ok {
				if err := u.UnmarshalWSV(field.Value, format); err != nil {
					return nil, &unmarshalError{
						tag:        name,
						attributes: keys,
						fieldIndex: fi.Index,
						fieldType:  sf.Type().String(),
						cause:      err,
					}
				}
				continue fl
			}
		}
		switch sf.Type().Kind() {
		case reflect.String:
			sf.SetString(field.Value)
			continue fl
		case reflect.Bool:
			v, err := internal.ParseBool(field.Value)
			if err != nil {
				return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: sf.Kind().String(), cause: err}
			}
			sf.SetBool(v)
			continue fl
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, err := strconv.ParseInt(field.Value, 10, strconv.IntSize)
			if err != nil {
				return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: sf.Kind().String(), cause: err}
			}
			sf.SetInt(v)
			continue fl
		case reflect.Float32, reflect.Float64:
			v, err := strconv.ParseFloat(field.Value, strconv.IntSize)
			if err != nil {
				return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: sf.Kind().String(), cause: err}
			}
			sf.SetFloat(v)
			continue fl
		case reflect.Ptr:
			if field.IsNull {
				continue fl
			}
			switch sf.Interface().(type) {
			case *time.Time:
				v, err := time.Parse(format, field.Value)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*time.Time", cause: err}
				}
				sf.Set(reflect.ValueOf(&v))
				continue fl
			case *string:
				sf.Set(reflect.ValueOf(&field.Value))
				continue fl
			case *int:
				v, err := strconv.ParseInt(field.Value, 10, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*int", cause: err}
				}
				x := int(v)
				sf.Set(reflect.ValueOf(&x))
				continue fl
			case *int8:
				v, err := strconv.ParseInt(field.Value, 10, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*int8", cause: err}
				}
				x := int8(v)
				sf.Set(reflect.ValueOf(&x))
				continue fl
			case *int16:
				v, err := strconv.ParseInt(field.Value, 10, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*int16", cause: err}
				}
				x := int16(v)
				sf.Set(reflect.ValueOf(&x))
				continue fl
			case *int32:
				v, err := strconv.ParseInt(field.Value, 10, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*int32", cause: err}
				}
				x := int32(v)
				sf.Set(reflect.ValueOf(&x))
				continue fl
			case *int64:
				v, err := strconv.ParseInt(field.Value, 10, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*int64", cause: err}
				}
				sf.Set(reflect.ValueOf(&v))
				continue fl
			case *float32:
				v, err := strconv.ParseFloat(field.Value, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*float32", cause: err}
				}
				x := float32(v)
				sf.Set(reflect.ValueOf(&x))
				continue fl
			case *float64:
				v, err := strconv.ParseFloat(field.Value, strconv.IntSize)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*float64", cause: err}
				}
				sf.Set(reflect.ValueOf(&v))
				continue fl
			case *bool:
				v, err := internal.ParseBool(field.Value)
				if err != nil {
					return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: "*bool", cause: err}
				}
				sf.Set(reflect.ValueOf(&v))
				continue fl
			default:
				return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: sf.Type().String()}
			}
		default:
			return nil, &unmarshalError{tag: name, attributes: keys, fieldIndex: fi.Index, fieldType: sf.Type().String()}
		}
	}

	return &val, nil
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
		keys := utils.SplitQuoted(tag)
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
