package validator

import "strings"

func (v *Validator) validate_bool(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return rtnErrs
	}

	switch strings.ToLower(strValue) {
	case "true", "false":
	default:
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
