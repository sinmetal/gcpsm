package ucon

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TagJSON provides methods to operate "json" tag.
type TagJSON string

// NewTagJSON returns new TagJSON.
func NewTagJSON(tag reflect.StructTag) TagJSON {
	return TagJSON(tag.Get("json"))
}

// Ignored returns whether json tag is ignored.
func (jsonTag TagJSON) Ignored() bool {
	return strings.Split(string(jsonTag), ",")[0] == "-"
}

// HasString returns whether a field is emitted as string.
func (jsonTag TagJSON) HasString() bool {
	for _, tag := range strings.Split(string(jsonTag), ",")[1:] {
		if tag == "string" {
			return true
		}
	}
	return false
}

// Name returns json tag name.
func (jsonTag TagJSON) Name() string {
	return strings.Split(string(jsonTag), ",")[0]
}

// OmitEmpty returns whether json tag is set as omitempty.
func (jsonTag TagJSON) OmitEmpty() bool {
	for _, tag := range strings.Split(string(jsonTag), ",")[1:] {
		if tag == "omitempty" {
			return true
		}
	}
	return false
}

func structFieldToKey(sf reflect.StructField) string {
	tagJSON := NewTagJSON(sf.Tag)
	if v := tagJSON.Name(); v != "" {
		return v
	}

	return sf.Name
}

func valueStringMapper(target reflect.Value, key string, value string) (bool, error) {
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	for i, numField := 0, target.NumField(); i < numField; i++ {
		sf := target.Type().Field(i)
		if NewTagJSON(sf.Tag).Ignored() {
			continue
		}

		f := target.Field(i)

		if sf.Anonymous {
			ret, err := valueStringMapper(f, key, value)
			if err != nil {
				return false, err
			}
			if ret {
				return true, nil
			}
			continue
		}

		if structFieldToKey(sf) != key {
			continue
		}

		if ft := f.Type(); ft.AssignableTo(stringParserType) {
			v, err := reflect.New(ft).Interface().(StringParser).ParseString(value)
			if err != nil {
				return true, err
			}
			f.Set(reflect.ValueOf(v))
			return true, nil
		}

		err := SetValueFromString(f, value)
		if err != nil {
			return true, err
		}

		return true, nil
	}

	return false, nil
}

func valueStringSliceMapper(target reflect.Value, key string, values []string) (bool, error) {
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	nf := target.NumField()
	for i := 0; i < nf; i++ {
		sf := target.Type().Field(i)
		if NewTagJSON(sf.Tag).Ignored() {
			continue
		}

		f := target.Field(i)

		if sf.Anonymous {
			ret, err := valueStringSliceMapper(f, key, values)
			if err != nil {
				return false, err
			}
			if ret {
				return true, nil
			}
			continue
		}

		if structFieldToKey(sf) != key {
			continue
		}

		ft := f.Type()
		if ft.Kind() != reflect.Slice {
			if len(values) == 0 {
				continue
			}
			ret, err := valueStringMapper(target, key, values[0])
			if err != nil {
				return false, err
			}
			if ret {
				return true, nil
			}
			continue
		}

		if fte := ft.Elem(); fte.AssignableTo(stringParserType) {
			sp := reflect.New(fte).Interface().(StringParser)
			for _, value := range values {
				v, err := sp.ParseString(value)
				if err != nil {
					return false, err
				}
				f.Set(reflect.Append(f, reflect.ValueOf(v)))
			}
			return true, nil
		}

		err := SetValueFromStrings(f, values)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// CheckFunction checks whether the target is a function.
func CheckFunction(target interface{}) {
	if reflect.ValueOf(target).Kind() != reflect.Func {
		panic("argument is not function")
	}
}

// IsEmpty returns whether the value is empty.
func IsEmpty(fV reflect.Value) bool {
	switch fV.Kind() {
	case reflect.Ptr, reflect.Interface:
		return fV.IsNil()
	case reflect.String:
		return fV.String() == ""
	case reflect.Array, reflect.Slice:
		return fV.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fV.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fV.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return fV.Float() == 0
	}

	return false
}

// SetValueFromString parses string and sets value.
func SetValueFromString(f reflect.Value, value string) error {
	ft := f.Type()
	if ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
		f = f.Elem()
	}

	switch ft.Kind() {
	case reflect.String:
		f.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			v, err := strconv.ParseInt(value, 0, ft.Bits())
			if err != nil {
				return newBadRequestf("%s is not int format", value)
			}
			f.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			v, err := strconv.ParseUint(value, 0, ft.Bits())
			if err != nil {
				return newBadRequestf("%s is not uint format", value)
			}
			f.SetUint(v)
		}
	case reflect.Float32, reflect.Float64:
		{
			v, err := strconv.ParseFloat(value, ft.Bits())
			if err != nil {
				return newBadRequestf("%s is not float format", value)
			}
			f.SetFloat(v)
		}
	case reflect.Bool:
		{
			v, err := strconv.ParseBool(value)
			if err != nil {
				return newBadRequestf("%s is not bool format", value)
			}
			f.SetBool(v)
		}
	case reflect.Slice, reflect.Array:
		elem := reflect.New(ft.Elem()).Elem()
		err := SetValueFromString(elem, value)
		if err != nil {
			return err
		}
		result := reflect.New(ft).Elem()
		result = reflect.Append(result, elem)
		f.Set(result)
	default:
		return fmt.Errorf("unsupported format %s", ft.Name())
	}

	return nil
}

// SetValueFromStrings parses strings and sets value.
func SetValueFromStrings(f reflect.Value, values []string) error {
	ft := f.Type()

	if ft.Kind() != reflect.Slice && len(values) == 1 {
		err := SetValueFromString(f, values[0])
		if err != nil {
			return err
		}
		return nil
	}

	switch el := ft.Elem(); el.Kind() {
	case reflect.String:
		f.Set(reflect.ValueOf(values))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			resultList := reflect.MakeSlice(ft, 0, len(values))
			for _, value := range values {
				v, err := strconv.ParseInt(value, 0, el.Bits())
				if err != nil {
					return newBadRequestf("%s is not int format", value)
				}

				resultList = reflect.Append(resultList, reflect.ValueOf(v).Convert(el))
			}
			f.Set(resultList)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			resultList := reflect.MakeSlice(ft, 0, len(values))
			for _, value := range values {
				v, err := strconv.ParseUint(value, 0, el.Bits())
				if err != nil {
					return newBadRequestf("%s is not uint format", value)
				}

				resultList = reflect.Append(resultList, reflect.ValueOf(v).Convert(el))
			}
			f.Set(resultList)
		}
	case reflect.Float32, reflect.Float64:
		{
			resultList := reflect.MakeSlice(ft, 0, len(values))
			for _, value := range values {
				v, err := strconv.ParseFloat(value, el.Bits())
				if err != nil {
					return newBadRequestf("%s is not float format", value)
				}

				resultList = reflect.Append(resultList, reflect.ValueOf(v).Convert(el))
			}
			f.Set(resultList)
		}
	case reflect.Bool:
		{
			resultList := make([]bool, 0, len(values))
			for _, value := range values {
				v, err := strconv.ParseBool(value)
				if err != nil {
					return newBadRequestf("%s is not bool format", value)
				}
				resultList = append(resultList, v)
			}
			f.Set(reflect.ValueOf(resultList))
		}
	default:
		return fmt.Errorf("unsupported format %s", el.Name())
	}

	return nil
}
