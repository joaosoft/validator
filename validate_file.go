package validator

import (
	"os"
	"reflect"
)

func (v *Validator) validate_file(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if _, err := os.Stat(v._convertToString(value)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
