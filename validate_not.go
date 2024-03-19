package validator

func (v *Validator) validate_not(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if strValue == v._convertToString(expected) {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
