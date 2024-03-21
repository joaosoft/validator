package validator

import "reflect"

func (v *Validator) validate_set_empty(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	obj.Set(reflect.Zero(reflect.TypeOf(value)))

	return rtnErrs
}
