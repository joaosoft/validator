package validator

import (
	"math/rand"
	"reflect"
)

func (v *Validator) validate_set_random(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	isPointer := false

	if kind == reflect.Ptr && !obj.IsNil() {
		isPointer = true
		value = obj.Elem()
		kind = reflect.TypeOf(obj.Interface()).Kind()
	}

	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if v._convertToString(expected) == "" {
			expected = value
		}

		if err = _setValue(kind, obj, v._random(v._convertToString(expected))); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min := 1
		max := 100
		obj.SetInt(int64(rand.Intn(max-min) + min))

	case reflect.Float32:
		obj.SetFloat(float64(rand.Float32()))

	case reflect.Float64:
		obj.SetFloat(rand.Float64())

	case reflect.Bool:
		obj.SetBool(rand.Intn(1) == 1)
	}

	if isPointer {
		value = obj.Addr()
	}

	return rtnErrs
}
