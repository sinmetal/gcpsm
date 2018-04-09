package golidator

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// ValidationResult provides result of validation.
type ValidationResult int

const (
	// ValidationNG means validation is failure.
	ValidationNG ValidationResult = iota
	// ValidationOK means validation is succeed.
	ValidationOK
)

// Validator is holder of validation information.
type Validator struct {
	tag   string
	funcs map[string]ValidationFunc
}

// ValidationFunc is validation function itself.
type ValidationFunc func(param string, value reflect.Value) (ValidationResult, error)

// NewValidator create and setup new Validator.
func NewValidator() *Validator {
	v := &Validator{}
	v.SetTag("validate")
	v.SetValidationFunc("req", ReqValidator)
	v.SetValidationFunc("d", DefaultValidator)
	v.SetValidationFunc("min", MinValidator)
	v.SetValidationFunc("max", MaxValidator)
	v.SetValidationFunc("minLen", MinLenValidator)
	v.SetValidationFunc("maxLen", MaxLenValidator)
	v.SetValidationFunc("email", EmailValidator)
	v.SetValidationFunc("enum", EnumValidator)
	return v
}

// SetTag is setup tag name in struct field tags.
func (vl *Validator) SetTag(tag string) {
	vl.tag = tag
}

// SetValidationFunc is setup tag name with ValidationFunc.
func (vl *Validator) SetValidationFunc(name string, vf ValidationFunc) {
	if vl.funcs == nil {
		vl.funcs = make(map[string]ValidationFunc)
	}
	vl.funcs[name] = vf
}

// Validate argument value.
func (vl *Validator) Validate(v interface{}) error {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	w := &walker{
		v: vl,
		report: &ErrorReport{
			Root: rv,
			Type: "https://github.com/favclip/golidator",
		},
		ParentFieldName: "",
		Root:            rv,
		Current:         rv,
	}
	if err := w.walkStruct(); err != nil {
		return err
	}

	if len(w.report.Details) != 0 {
		return w.report
	}

	return nil
}

type walker struct {
	v      *Validator
	report *ErrorReport

	ParentFieldName string
	Root            reflect.Value
	Current         reflect.Value
}

func (w *walker) walkStruct() error {
	sv := w.Current
	st := sv.Type()

	for sv.Kind() == reflect.Ptr && !sv.IsNil() {
		sv = sv.Elem()
		st = sv.Type()
	}

	if sv.Kind() != reflect.Struct {
		return ErrUnsupportedValue
	}

	for i := 0; i < sv.NumField(); i++ {
		fv := sv.Field(i)
		ft := st.Field(i)
		for fv.Kind() == reflect.Ptr && !fv.IsNil() {
			fv = fv.Elem()
		}
		if !unicode.IsUpper([]rune(ft.Name)[0]) {
			// private field!
			continue
		}

		tag := ft.Tag.Get(w.v.tag)

		if tag == "-" {
			continue
		}
		if tag == "" {
			if fv.Kind() == reflect.Struct {
				if err := w.walkField(ft, fv); err != nil {
					return err
				}
			}
			continue
		}

		params, err := parseTag(tag)
		if err != nil {
			return err
		}

		err = w.validateField(ft, fv, params)
		if err != nil {
			return err
		}

		if fv.Kind() == reflect.Struct {
			if ft.Anonymous {
				if err := w.walkFieldWithParentFieldName(w.ParentFieldName, ft, fv); err != nil {
					return err
				}
			} else {
				if err := w.walkField(ft, fv); err != nil {
					return err
				}
			}
			continue
		}
	}

	return nil
}

func (w *walker) walkFieldWithParentFieldName(parentFieldName string, ft reflect.StructField, fv reflect.Value) error {
	if fv.Kind() != reflect.Struct {
		return ErrUnsupportedValue
	}

	w2 := &walker{
		v:               w.v,
		report:          w.report,
		ParentFieldName: parentFieldName,
		Root:            w.Root,
		Current:         fv,
	}
	return w2.walkStruct()
}

func (w *walker) walkField(ft reflect.StructField, fv reflect.Value) error {
	if fv.Kind() != reflect.Struct {
		return ErrUnsupportedValue
	}

	name := w.fieldName(ft)
	parentFieldName := w.ParentFieldName
	if parentFieldName != "" {
		parentFieldName += "."
	}
	parentFieldName += name

	return w.walkFieldWithParentFieldName(parentFieldName, ft, fv)
}

func (w *walker) validateField(ft reflect.StructField, fv reflect.Value, params map[string]string) error {
	var detail *ErrorDetail
	for k, v := range params {
		f, ok := w.v.funcs[k]
		if !ok {
			return fmt.Errorf("%s: unknown rule %s in %s", w.Current.Type().Name(), k, ft.Name)
		}
		result, err := f(v, fv)
		if err != nil {
			return err
		}
		if result == ValidationOK {
			continue
		}

		if detail == nil {
			name := w.fieldName(ft)
			if parent := w.ParentFieldName; parent != "" {
				name = parent + "." + name
			}
			detail = &ErrorDetail{
				Current:         w.Current,
				ParentFieldName: w.ParentFieldName,
				Value:           fv,
				Field:           ft,
				FieldName:       name,
			}
		}
		detail.ReasonList = append(detail.ReasonList, &ErrorReason{
			Type:   k,
			Config: v,
		})
	}
	if detail != nil {
		w.report.Details = append(w.report.Details, detail)
	}
	return nil
}

func (w *walker) fieldName(ft reflect.StructField) string {
	if tag := ft.Tag.Get("json"); tag != "" {
		vs := strings.SplitN(tag, ",", 2)
		if v := vs[0]; v != "" && v != "-" {
			return v
		}
	}

	return ft.Name
}

func parseTag(tagBody string) (map[string]string, error) {
	result := make(map[string]string, 0)
	if tagBody == "" {
		return result, nil
	}
	ss := strings.Split(tagBody, ",")
	for _, s := range ss {
		if s == "" {
			continue
		}
		p := strings.SplitN(s, "=", 2)
		name := p[0]
		if name == "" {
			return nil, ErrEmptyValidationName
		}
		if len(p) == 1 {
			result[name] = ""
			continue
		}

		result[name] = p[1]
	}

	return result, nil
}
