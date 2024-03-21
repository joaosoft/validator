package validator

import (
	"crypto/md5"
	"fmt"
	"reflect"
)

func (v *Validator) validate_set_md5(context *ValidatorContext, validationData *ValidationData) []error {
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

		newValue := fmt.Sprintf("%x", md5.Sum([]byte(v._convertToString(expected))))
		if err = _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}
