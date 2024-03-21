package validator

import (
	"reflect"
	"strings"
)

func (v *Validator) validate_suffix(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !strings.HasSuffix(v._convertToString(value), v._convertToString(expected)) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
