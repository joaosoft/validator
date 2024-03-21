package validator

import (
	"regexp"
	"unicode/utf8"
)

func (v *Validator) validate_regex(context *ValidatorContext, validationData *ValidationData) []error {

	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	r, err := regexp.Compile(validationData.Expected.(string))
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if utf8.RuneCountInString(strValue) > 0 {
		if !r.MatchString(strValue) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
