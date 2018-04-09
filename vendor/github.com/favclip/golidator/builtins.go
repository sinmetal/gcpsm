package golidator

import (
	"bytes"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ReqValidator check value that must not be empty.
func ReqValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return ValidationNG, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() == 0 {
			return ValidationNG, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.Uint() == 0 {
			return ValidationNG, nil
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() == 0 {
			return ValidationNG, nil
		}
	case reflect.Bool:
	// :)

	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			return ValidationNG, nil
		}
	case reflect.Struct:
		// :)
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return ValidationNG, nil
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// DefaultValidator set value when value is empty.
func DefaultValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.String:
		if !v.CanAddr() {
			return ValidationNG, nil
		}
		if v.String() == "" {
			v.SetString(param)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if !v.CanAddr() {
			return ValidationNG, nil
		}
		pInt, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return ValidationNG, nil
		}
		if v.Int() == 0 {
			v.SetInt(pInt)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if !v.CanAddr() {
			return ValidationNG, nil
		}
		pUint, err := strconv.ParseUint(param, 0, 64)
		if err != nil {
			return ValidationNG, nil
		}
		if v.Uint() == 0 {
			v.SetUint(pUint)
		}
	case reflect.Float32, reflect.Float64:
		if !v.CanAddr() {
			return ValidationNG, nil
		}
		pFloat, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return ValidationNG, nil
		}
		if v.Float() == 0 {
			v.SetFloat(pFloat)
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// MinValidator check value that must greater or equal than config value.
func MinValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		pInt, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if v.Int() < pInt {
			return ValidationNG, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		pUint, err := strconv.ParseUint(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if v.Uint() < pUint {
			return ValidationNG, nil
		}
	case reflect.Float32, reflect.Float64:
		pFloat, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if v.Float() < pFloat {
			return ValidationNG, nil
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// MaxValidator check value that must less or equal than config value.
func MaxValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		pInt, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if v.Int() > pInt {
			return ValidationNG, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		pUint, err := strconv.ParseUint(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if v.Uint() > pUint {
			return ValidationNG, nil
		}
	case reflect.Float32, reflect.Float64:
		pFloat, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if v.Float() > pFloat {
			return ValidationNG, nil
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// MinLenValidator check value length that must greater or equal than config value.
func MinLenValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return ValidationOK, nil // emptyの場合は無視 これはreqの役目だ
		}
		p, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if int64(utf8.RuneCountInString(v.String())) < p {
			return ValidationNG, nil
		}
	case reflect.Array, reflect.Map, reflect.Slice:
		if v.Len() == 0 {
			return ValidationOK, nil
		}
		p, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if int64(v.Len()) < p {
			return ValidationNG, nil
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// MaxLenValidator check value length that must less or equal than config value.
func MaxLenValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return ValidationOK, nil // emptyの場合は無視 これはreqの役目だ
		}
		p, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if int64(utf8.RuneCountInString(v.String())) > p {
			return ValidationNG, nil
		}
	case reflect.Array, reflect.Map, reflect.Slice:
		if v.Len() == 0 {
			return ValidationOK, nil
		}
		p, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrInvalidConfigValue
		}
		if int64(v.Len()) > p {
			return ValidationNG, nil
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// atext in RFC5322
// http://www.hde.co.jp/rfc/rfc5322.php?page=12
var atextChars = []byte(
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"!#$%&'*+-/=?^_`{|}~")

// isAtext is return atext contains `c`
func isAtext(c rune) bool {
	return bytes.IndexRune(atextChars, c) >= 0
}

// EmailValidator check value that must be email address format.
func EmailValidator(param string, v reflect.Value) (ValidationResult, error) {
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return ValidationOK, nil
		}
		addr := v.String()
		// do validation by RFC5322
		// http://www.hde.co.jp/rfc/rfc5322.php?page=17

		// screening
		if _, err := mail.ParseAddress(addr); err != nil {
			return ValidationNG, nil
		}

		addrSpec := strings.Split(addr, "@")
		if len(addrSpec) != 2 {
			return ValidationNG, nil
		}
		// check local part
		localPart := addrSpec[0]
		// divided by quoted-string style or dom-atom style
		if match, err := regexp.MatchString(`"[^\t\n\f\r\\]*"`, localPart); err == nil && match { // "\"以外の表示可能文字を認める
			// OK
		} else if match, err := regexp.MatchString(`^([^.\s]+\.)*([^.\s]+)$`, localPart); err != nil || !match { // (hoge.)*hoge
			return ValidationNG, nil
		} else {
			// atext check for local part
			for _, c := range localPart {
				if string(c) == "." {
					// "." is already checked by regexp
					continue
				}
				if !isAtext(c) {
					return ValidationNG, nil
				}
			}
		}
		// check domain part
		domain := addrSpec[1]
		if match, err := regexp.MatchString(`^([^.\s]+\.)*[^.\s]+$`, domain); err != nil || !match { // (hoge.)*hoge
			return ValidationNG, nil
		}
		// atext check for domain part
		for _, c := range domain {
			if string(c) == "." {
				// "." is already checked by regexp
				continue
			}
			if !isAtext(c) {
				return ValidationNG, nil
			}
		}
	default:
		return 0, ErrValidateUnsupportedType
	}

	return ValidationOK, nil
}

// EnumValidator check value that must contains in config values.
func EnumValidator(param string, v reflect.Value) (ValidationResult, error) {
	if param == "" {
		return 0, ErrInvalidConfigValue
	}

	params := strings.Split(param, "|")

	var enum func(v reflect.Value) (ValidationResult, error)
	enum = func(v reflect.Value) (ValidationResult, error) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.String:
			val := v.String()
			if val == "" {
				// need empty checking? use req :)
				return ValidationOK, nil
			}
			for _, value := range params {
				if val == value {
					return ValidationOK, nil
				}
			}
			return ValidationNG, nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := v.Int()
			for _, value := range params {
				value2, err := strconv.ParseInt(value, 10, 0)
				if err != nil {
					return 0, ErrInvalidConfigValue
				}
				if val == value2 {
					return ValidationOK, nil
				}
			}
			return ValidationNG, nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			val := v.Uint()
			for _, value := range params {
				value2, err := strconv.ParseUint(value, 10, 0)
				if err != nil {
					return 0, ErrInvalidConfigValue
				}
				if val == value2 {
					return ValidationOK, nil
				}
			}
			return ValidationNG, nil
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				e := v.Index(i)
				return enum(e)
			}
		default:
			return 0, ErrValidateUnsupportedType
		}

		return ValidationOK, nil
	}

	return enum(v)
}
