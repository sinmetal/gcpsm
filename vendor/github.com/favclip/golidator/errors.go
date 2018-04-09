package golidator

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

// ErrEmptyValidationName means invalid validator name error.
var ErrEmptyValidationName = errors.New("validator name is required")

// ErrUnsupportedValue means golidator is not support type of passed value.
var ErrUnsupportedValue = errors.New("unsupported type")

// ErrValidateUnsupportedType means validator is not support field type.
var ErrValidateUnsupportedType = errors.New("unsupported field type")

// ErrInvalidConfigValue means config value can't accept by validator.
var ErrInvalidConfigValue = errors.New("invalid configuration value")

// ErrorReport provides detail of error informations.
type ErrorReport struct {
	Root reflect.Value `json:"-"`

	Type    string         `json:"type"` // fixed value. https://github.com/favclip/golidator
	Details []*ErrorDetail `json:"details"`
}

// ErrorDetail provides error about 1 field.
type ErrorDetail struct {
	ParentFieldName string              `json:"-"`
	Current         reflect.Value       `json:"-"`
	Value           reflect.Value       `json:"-"`
	Field           reflect.StructField `json:"-"`

	FieldName  string         `json:"fieldName"` // e.g. address , person.name
	ReasonList []*ErrorReason `json:"reasonList"`
}

// ErrorReason contains why validation is failed?
type ErrorReason struct {
	Type   string `json:"type"`             // e.g. req , min, enum
	Config string `json:"config,omitempty"` // e.g. 1, manual|auto, ^[^@]+@gmail.com$
}

func (report *ErrorReport) Error() string {
	buf := bytes.NewBufferString("invalid. ")

	for t := report.Root.Type(); ; {
		switch t.Kind() {
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
			t = t.Elem()
			continue
		}
		if name := t.Name(); name != "" {
			fmt.Fprint(buf, name, " ")
		}
		break
	}
	for idx, detail := range report.Details {
		fmt.Fprint(buf, "#", idx+1, " ", detail.FieldName, ": ")
		for _, report := range detail.ReasonList {
			fmt.Fprint(buf, report.Type)
			if report.Config != "" {
				fmt.Fprint(buf, "=", report.Config)
			}
			if detail.Value.Kind() == reflect.String {
				fmt.Fprintf(buf, " actual: '%v'", detail.Value.Interface())
			} else {
				fmt.Fprintf(buf, " actual: %v", detail.Value.Interface())
			}
		}
		if idx != len(report.Details)-1 {
			fmt.Fprint(buf, ", ")
		}
	}

	return buf.String()
}
