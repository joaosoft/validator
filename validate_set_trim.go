package validator

import (
	"bytes"
	"reflect"
	"regexp"
	"strings"
)

func (v *Validator) validate_set_trim(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.TrimSpace(value.(string))

		r, err := regexp.Compile(constRegexForTrim)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		newValue = string(r.ReplaceAll(bytes.TrimSpace([]byte(newValue)), []byte(" ")))
		if err = _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}
