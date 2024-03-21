package validator

import (
	"reflect"
	"strings"
)

func (v *Validator) validate_set_lower(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.ToLower(value.(string))
		if err := _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}
