package validator

import "unicode"

func (v *Validator) validate_numeric(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return rtnErrs
	}

	for _, r := range strValue {
		if !unicode.IsNumber(r) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
			break
		}
	}

	return rtnErrs
}
