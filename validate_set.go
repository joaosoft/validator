package validator

import "reflect"

func (v *Validator) validate_set(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	if err = _setValue(kind, obj, expected); err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	return rtnErrs
}
