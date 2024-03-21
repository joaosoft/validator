package validator

import (
	uuid "github.com/satori/go.uuid"
	"reflect"
)

func (v *Validator) validate_uuid(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)
	check := false

	_, obj, value := v._getValue(validationData.Value)

	var checkValue interface{}
	switch obj.Type() {
	case reflect.TypeOf(uuid.UUID{}):
		check = true
		checkValue = obj.Interface().(uuid.UUID).String()
	case reflect.TypeOf(""):
		check = true
		checkValue = value
	}

	if check {
		if _, err := uuid.FromString(v._convertToString(checkValue)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
