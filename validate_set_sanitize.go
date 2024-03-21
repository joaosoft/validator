package validator

import "strings"

func (v *Validator) validate_set_sanitize(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	split := strings.Split(validationData.Expected.(string), constTagSplitValues)
	invalid := make([]string, 0)

	// validate expected
	for _, str := range split {
		if strings.Contains(strValue, str) {
			invalid = append(invalid, str)
		}
	}

	// validate global
	for _, str := range v.sanitize {
		if strings.Contains(strValue, str) {
			invalid = append(invalid, str)
		}
	}

	if len(invalid) > 0 {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
