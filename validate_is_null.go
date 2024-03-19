package validator

func (v *Validator) validate_is_null(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, _ := v._getValue(validationData.Value)
	if !isNil {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
