package validator

import (
	"encoding/base64"
	"reflect"
)

func (v *Validator) validate_base64(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if _, err := base64.StdEncoding.DecodeString(v._convertToString(value)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
