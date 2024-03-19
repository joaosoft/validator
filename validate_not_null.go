package validator

func (v *Validator) validate_not_null(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if errs := v.validate_is_null(context, validationData); len(errs) == 0 {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
