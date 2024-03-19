package validator

import (
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (v *Validator) validate_size(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, obj, _ := v._getValue(validationData.Value)
	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	size, e := strconv.Atoi(v._convertToString(expected))
	if e != nil {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
		return rtnErrs
	}

	var valueSize int64

	switch obj.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(obj.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = int64(utf8.RuneCountInString(strings.TrimSpace(strconv.Itoa(int(obj.Int())))))
	case reflect.Float32, reflect.Float64:
		valueSize = int64(utf8.RuneCountInString(strings.TrimSpace(strconv.FormatFloat(obj.Float(), 'g', 1, 64))))
	case reflect.String:
		valueSize = int64(utf8.RuneCountInString(strings.TrimSpace(obj.String())))
	case reflect.Bool:
		valueSize = int64(utf8.RuneCountInString(strings.TrimSpace(strconv.FormatBool(obj.Bool()))))
	default:
		if isNil {
			break
		}
		valueSize = int64(utf8.RuneCountInString(strings.TrimSpace(obj.String())))
	}

	if valueSize != int64(size) {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
