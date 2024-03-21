package validator

func (v *Validator) validate_password(context *ValidatorContext, validationData *ValidationData) (errs []error) {
	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return nil
	}

	return v.password.settings.Compare(strValue)
}
