package validator

import (
	"reflect"
	"strings"
	"unicode/utf8"

	uuid "github.com/satori/go.uuid"
)

func (v *Validator) validate_is_empty(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	var isZero bool

	isNil, obj, value := v._getValue(validationData.Value)

	switch obj.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:

		switch obj.Type() {
		case reflect.TypeOf(uuid.UUID{}):
			if value.(uuid.UUID) == uuid.Nil {
				isZero = true
			}
		default:
			isZero = obj.Len() == 0
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		isZero = obj.Int() == 0
	case reflect.Float32, reflect.Float64:
		isZero = obj.Float() == 0
	case reflect.String:
		isZero = utf8.RuneCountInString(strings.TrimSpace(obj.String())) == 0
	case reflect.Bool:
		isZero = obj.Bool() == false
	case reflect.Struct:
		if reflect.DeepEqual(value, reflect.New(obj.Type()).Interface()) {
			isZero = true
		}
	default:
		if isNil {
			isZero = true
		}
	}

	if !isZero {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
