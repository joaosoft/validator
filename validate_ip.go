package validator

import (
	"net"
	"reflect"
)

func (v *Validator) validate_ip(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if ip := net.ParseIP(v._convertToString(value)); ip == nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
