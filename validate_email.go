package validator

import (
	"reflect"
	"regexp"
)

func (v *Validator) validate_email(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		r, err := regexp.Compile(constRegexForEmail)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !r.MatchString(v._convertToString(value)) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
