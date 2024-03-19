package validator

import (
	"reflect"
)

func (v *Validator) validate_id(context *ValidatorContext, validationData *ValidationData) []error {
	id := v._convertToString(validationData.Expected)
	dat := &data{
		value: validationData.Value,
		typ: reflect.StructField{
			Type: reflect.TypeOf(validationData.Value),
		},
	}
	context.SetValue(constTagId, id, dat)

	return nil
}
