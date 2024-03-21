package validator

import (
	"reflect"
	"strings"
)

func (v *Validator) validate_set_key(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if expected == "" {
			expected = value
		}

		if err = _setValue(kind, obj, convertToKey(strings.TrimSpace(v._convertToString(expected)), true)); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}
